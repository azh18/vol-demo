package main

import (
	"encoding/json"
	"fmt"
	"github.com/bytedance/sonic"
	pb "github.com/zbw0046/vol-demo/grpc"
	"github.com/zbw0046/vol-demo/utils"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

type backendServer struct {
	pb.UnimplementedBackendServer
}

func (b *backendServer) WatchUpload(request *pb.WatchUploadRequest, server pb.Backend_WatchUploadServer) error {
	vid := request.GetVid()
	log.Printf("WatchUpload vid=%s", vid)
	watcher, err := GetService().Watch(vid)
	if err != nil {
		return err
	}
	defer watcher.Close()

	for msg := range watcher.ResultChan() {
		if err := server.Send(&pb.WatchUploadResponse{Message: msg}); err != nil {
			log.Printf("grpc send msg %s error: %s", msg, err.Error())
		}
	}
	return nil
}

func invokeUpload(writer http.ResponseWriter, request *http.Request) {
	vid, err := upload(request.Body)
	if err != nil {
		log.Printf("invokeUpload error: %s", err.Error())
		utils.JSON(writer, http.StatusInternalServerError, err.Error())
		return
	}
	_, _ = writer.Write([]byte(vid))
}

func upload(data io.ReadCloser) (vid string, err error) {
	baseDir := "/tmp/volc_demo"
	if err = os.MkdirAll(baseDir, 0777); err != nil {
		return
	}
	filePath := filepath.Join(baseDir, utils.RandString(32))
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	if _, err = io.Copy(f, data); err != nil {
		return
	}

	vid, err = GetVolcSDK().UploadFile(filePath)
	if err != nil {
		return
	}

	GetService().OnUploadSuccess(vid, nil)
	return
}

type VolCallbackEvent struct {
	RequestID string          `json:"RequestId"`
	Version   string          `json:"Version"`
	EventType string          `json:"EventType"`
	EventTime string          `json:"EventTime"`
	Data      json.RawMessage `json:"Data"`
}

type WorkflowCompleteData struct {
	Code         string `json:"Code"`
	VID          string `json:"Vid"`
	CallbackArgs string `json:"CallbackArgs"`
	SpaceName    string `json:"SpaceName"`
	Message      string `json:"Message"`
	RunId        string `json:"RunId"`
	TemplateId   string `json:"TemplateId"`
}

func callbackHandler(writer http.ResponseWriter, request *http.Request) {
	// only process transcode finish callback
	event := &VolCallbackEvent{}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("read body error: %s", err.Error())
		return
	}

	if err := sonic.Unmarshal(body, event); err != nil {
		errStr := fmt.Sprintf("unmarshal request to event error: %s", err.Error())
		log.Printf(errStr)
		utils.JSON(writer, http.StatusBadRequest, errStr)
		return
	}

	if event.EventType != "WorkflowComplete" {
		writer.WriteHeader(http.StatusOK)
		return
	}

	workflowCompleteData := &WorkflowCompleteData{}
	if err := sonic.Unmarshal(event.Data, workflowCompleteData); err != nil {
		errStr := fmt.Sprintf("unmarshal request to workflowCompleteData error: %s", err.Error())
		log.Printf(errStr)
		utils.JSON(writer, http.StatusBadRequest, errStr)
		return
	}

	GetService().OnTranscodeFinish(workflowCompleteData.VID, workflowCompleteData)
	writer.WriteHeader(http.StatusOK)
}

func main() {
	go runGRPCServer()
	runUploadServer()
}

func runUploadServer() {
	m := http.NewServeMux()
	m.HandleFunc("/upload", invokeUpload)
	m.HandleFunc("/callback", callbackHandler)
	if err := http.ListenAndServe(":8081", m); err != nil {
		log.Printf("ListenAndServe on upload server return %s", err.Error())
	}
}

func runGRPCServer() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterBackendServer(s, &backendServer{})

	if err := s.Serve(l); err != nil {
		log.Printf("ListenAndServe on gRPC server return %s", err.Error())
	}
}
