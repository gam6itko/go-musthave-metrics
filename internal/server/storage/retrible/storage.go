package retrible

import (
	"errors"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
	"time"
)

type Error struct {
	inner error
}

func NewError(inner error) *Error {
	return &Error{
		inner,
	}
}

// Error добавляет поддержку интерфейса error для типа TimeError.
func (ths Error) Error() string {
	return fmt.Sprintf("%v", ths.inner)
}

func (ths Error) Unwrap() error {
	return ths.inner
}

type Storage struct {
	inner   storage.Storage
	tryEach []time.Duration
}

func NewStorage(inner storage.Storage, tryEach []time.Duration) *Storage {
	tryEach = append(tryEach, time.Nanosecond)
	return &Storage{
		inner,
		tryEach,
	}
}

func (ths Storage) CounterInc(name string, val int64) (err error) {
	for _, d := range ths.tryEach {
		err = ths.inner.CounterInc(name, val)
		if err == nil {
			return
		}

		if errors.Is(err, Error{}) {
			time.Sleep(d)
		}
	}

	return
}

func (ths Storage) CounterGet(name string) (result int64, err error) {
	for _, d := range ths.tryEach {
		result, err = ths.inner.CounterGet(name)
		if err == nil {
			return
		}

		if errors.Is(err, Error{}) {
			time.Sleep(d)
		}
	}

	return
}

func (ths Storage) CounterAll() (result map[string]int64, err error) {
	for _, d := range ths.tryEach {
		result, err = ths.inner.CounterAll()
		if err == nil {
			return
		}

		if errors.Is(err, Error{}) {
			time.Sleep(d)
		}
	}

	return
}

func (ths Storage) GaugeSet(name string, val float64) (err error) {
	for _, d := range ths.tryEach {
		err = ths.inner.GaugeSet(name, val)
		if err == nil {
			return
		}

		if errors.Is(err, Error{}) {
			time.Sleep(d)
		}
	}

	return
}

func (ths Storage) GaugeGet(name string) (result float64, err error) {
	for _, d := range ths.tryEach {
		result, err = ths.inner.GaugeGet(name)
		if err == nil {
			return
		}

		if errors.Is(err, Error{}) {
			time.Sleep(d)
		}
	}

	return
}
func (ths Storage) GaugeAll() (result map[string]float64, err error) {
	for _, d := range ths.tryEach {
		result, err = ths.inner.GaugeAll()
		if err == nil {
			return
		}

		if errors.Is(err, Error{}) {
			time.Sleep(d)
		}
	}

	return
}
