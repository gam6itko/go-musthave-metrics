package file

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
	"io"
	"os"
)

// Storage decorator on memory.Storage
type Storage struct {
	inner storage.IStorage
	file  *os.File
}

func NewStorage(inner storage.IStorage, filepath string, ioSync bool) (*Storage, error) {
	flag := os.O_RDWR | os.O_CREATE
	if ioSync {
		flag |= os.O_SYNC
	}
	file, err := os.OpenFile(filepath, flag, 0774)
	if err != nil {
		return nil, err
	}

	return &Storage{
		inner,
		file,
	}, nil
}

func (ths Storage) GaugeSet(ctx context.Context, name string, val float64) error {
	if err := ths.inner.GaugeSet(ctx, name, val); err != nil {
		return err
	}
	return ths.Save()
}

func (ths Storage) GaugeGet(ctx context.Context, name string) (float64, error) {
	return ths.inner.GaugeGet(ctx, name)
}

func (ths Storage) GaugeAll(ctx context.Context) (map[string]float64, error) {
	return ths.inner.GaugeAll(ctx)
}

func (ths Storage) CounterInc(ctx context.Context, name string, val int64) error {
	if err := ths.inner.CounterInc(ctx, name, val); err != nil {
		return err
	}
	return ths.Save()
}

func (ths Storage) CounterGet(ctx context.Context, name string) (int64, error) {
	return ths.inner.CounterGet(ctx, name)
}

func (ths Storage) CounterAll(ctx context.Context) (map[string]int64, error) {
	return ths.inner.CounterAll(ctx)
}

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
		return fmt.Errorf("file seek error: %s", err)
	}

	fi, err := ths.file.Stat()
	if err != nil {
		return err
	}

	if fi.Size() > 0 {
		decoder := json.NewDecoder(ths.file)
		if err := decoder.Decode(&ths.inner); err != nil {
			return err
		}
	}

	return nil
}

func (ths Storage) Close() error {
	return ths.file.Close()
}
