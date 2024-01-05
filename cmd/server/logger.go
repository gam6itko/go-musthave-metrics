package main

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	Log = logger
}
