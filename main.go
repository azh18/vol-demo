package main

import (
	"go.uber.org/zap"
	"net/http"
)

var (
	logger *zap.SugaredLogger
)

func main() {
	// init logger
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l.Sugar()
	defer l.Sync()

	m := http.NewServeMux()
	m.HandleFunc("/callback", Callback)

	if err := http.ListenAndServe(":8888", m); err != nil {
		logger.Infof("ListenAndServe return %s", err.Error())
		return
	}
}
