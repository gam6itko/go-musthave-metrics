package main

import (
	"go.uber.org/zap"
	"log"
)

var Log *zap.Logger = zap.NewNop()

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("ERROR. failed to sync logger. %s", err)
		}
	}()

	Log = logger
}
