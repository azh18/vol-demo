package utils

import (
	"github.com/bytedance/sonic"
	"log"
	"net/http"
)

func JSON(writer http.ResponseWriter, code int, v interface{}) {
	writer.WriteHeader(code)
	body, err := sonic.Marshal(v)
	if err != nil {
		log.Printf("marshal error: %s", err.Error())
		return
	}
	if _, err = writer.Write(body); err != nil {
		log.Printf("write body error: %s", err.Error())
	}
}
