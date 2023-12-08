package flags

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port uint16
}

func NewNetAddr(host string, port uint16) NetAddress {
	return NetAddress{host, port}
}

func (ths *NetAddress) String() string {
	var portStr string
	if ths.Port > 0 {
		portStr = fmt.Sprintf(":%d", ths.Port)
	}
	return ths.Host + portStr
}

func (ths *NetAddress) Set(flagValue string) error {
	return ths.FromString(flagValue)
}

func (ths *NetAddress) FromString(dsn string) error {
	slice := strings.Split(dsn, ":")
	switch len(slice) {
	case 1:
		ths.Host = slice[0]
	case 2:
		ths.Host = slice[0]
		if port, err := strconv.ParseUint(slice[1], 10, 16); err != nil {
			return err
		} else {
			ths.Port = uint16(port)
		}
	default:
		return errors.New("invalid attr count")
	}

	return nil
}
