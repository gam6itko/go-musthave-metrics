package update

import (
	"errors"
	"github.com/gam6itko/go-musthave-metrics/internal/storage/memory"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

type PathMatcher struct {
	regExp *regexp.Regexp
}

func (ths PathMatcher) Match(path string) (string, string, string, error) {
	if !ths.regExp.MatchString(path) {
		return "", "", "", errors.New("not match")
	}

	match := ths.regExp.FindStringSubmatch(path)
	return match[1], match[2], match[3], nil
}

var pathMatcher PathMatcher

func init() {
	pathMatcher = PathMatcher{
		regexp.MustCompile(`^/update/(?P<type>\w+)/(?P<name>\w+)/(?P<value>\d+(?:\.\d+)?)$`),
	}
}

func Handler(resp http.ResponseWriter, req *http.Request) {
	mType, name, value, err := pathMatcher.Match(req.URL.Path)
	if err != nil {
		http.Error(resp, err.Error(), 404)
		return
	}

	switch mType {
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(resp, "invalid counter value", http.StatusBadRequest)
			return
		}
		memory.CounterInc(name, v)
	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(resp, "invalid gauge value", http.StatusBadRequest)
			return
		}
		memory.GaugeSet(name, v)
	default:
		http.Error(resp, "invalid metric type", http.StatusBadRequest)
		return
	}

	io.WriteString(resp, "OK")
	resp.WriteHeader(http.StatusOK)
}
