package flags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port uint
}

func NewNetAddr(host string, port uint) NetAddress {
	return NetAddress{host, port}
}

func (o *NetAddress) String() string {
	var portStr string
	if o.Port > 0 {
		portStr = fmt.Sprintf(":%d", o.Port)
	}
	return o.Host + portStr
}

func (o *NetAddress) Set(flagValue string) error {
	slice := strings.Split(flagValue, ":")
	switch len(slice) {
	case 1:
		o.Host = slice[0]
	case 2:
		o.Host = slice[0]
		if port, err := strconv.ParseInt(slice[1], 10, 16); err != nil {
			return err
		} else {
			o.Port = uint(port)
		}
	default:
		return errors.New("invalid attr count")
	}

	return nil
}
