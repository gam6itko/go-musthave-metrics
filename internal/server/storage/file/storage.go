package file

import (
	"encoding/json"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"io"
	"os"
)

// Storage decorator on memory.Storage
type Storage struct {
	inner *memory.Storage
	file  *os.File
}

func NewStorage(inner *memory.Storage, filepath string, sync bool) *Storage {
	flag := os.O_RDWR | os.O_CREATE
	if sync {
		flag |= os.O_SYNC
	}
	file, err := os.OpenFile(filepath, flag, 0774)
	if err != nil {
		panic(err)
	}

	return &Storage{
		inner,
		file,
	}
}

//<editor-fold desc="IMetricStorage decorator">

func (ths Storage) GaugeSet(name string, val float64) {
	ths.inner.GaugeSet(name, val)
	ths.Save()
}

func (ths Storage) GaugeGet(name string) (float64, bool) {
	return ths.inner.GaugeGet(name)
}

func (ths Storage) GaugeAll() map[string]float64 {
	return ths.inner.GaugeAll()
}

func (ths Storage) CounterInc(name string, val int64) {
	ths.inner.CounterInc(name, val)
	ths.Save()
}

func (ths Storage) CounterGet(name string) (int64, bool) {
	return ths.inner.CounterGet(name)
}

func (ths Storage) CounterAll() map[string]int64 {
	return ths.inner.CounterAll()
}

//</editor-fold>

func (ths Storage) Save() error {
	b, err := json.Marshal(ths.inner)
	if err != nil {
		return err
	}

	_, err = ths.file.WriteAt(b, 0)
	if err != nil {
		return err
	}

	return nil
}

func (ths Storage) Load() error {
	if _, err := ths.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	decoder := json.NewDecoder(ths.file)
	if err := decoder.Decode(&ths.inner); err != nil {
		return err
	}

	return nil
}

func (ths Storage) Close() error {
	return ths.file.Close()
}
