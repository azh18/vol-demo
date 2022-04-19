package main

import (
	"fmt"
	"github.com/bytedance/sonic"
	"io/ioutil"
	"net/http"
)

func Callback(writer http.ResponseWriter, request *http.Request) {
	req := &VolCallbackEvent{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		logger.Errorf("read body error: %s", err.Error())
		return
	}

	if err := sonic.Unmarshal(body, req); err != nil {
		errStr := fmt.Sprintf("unmarshal request to event error: %s", err.Error())
		logger.Errorf(errStr)
		JSON(writer, http.StatusBadRequest, errStr)
		return
	}
	logger.Infof("callback type: %s, data: %s", req.EventType, string(req.Data))
	JSON(writer, http.StatusOK, "ok")
}

func JSON(writer http.ResponseWriter, code int, v interface{}) {
	writer.WriteHeader(code)
	body, err := sonic.Marshal(v)
	if err != nil {
		logger.Errorf("marshal error: %s", err.Error())
		return
	}
	if _, err = writer.Write(body); err != nil {
		logger.Errorf("write body error: %s", err.Error())
	}
}
