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
		if err2 := logger.Sync(); err2 != nil {
			log.Fatal("failed to sync logger", err2)
		}
	}()

	Log = logger
}
