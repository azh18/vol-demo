package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/zbw0046/vol-demo/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	httpCli    = &http.Client{}
	serverHost = "127.0.0.1"
	gRPCConn   *grpc.ClientConn
)

func main() {
	filePath := ""
	vid := ""
	flag.StringVar(&filePath, "upload_path", "", "upload file path")
	flag.StringVar(&vid, "query_vid", "", "query vid")
	flag.StringVar(&serverHost, "server", "", "server host")
	flag.Parse()

	// init grpc client
	conn, err := grpc.Dial(fmt.Sprintf("%s:8080", serverHost), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	gRPCConn = conn

	if vid != "" {
		QueryVID(vid)
		return
	}
	RunUploadProcess(filePath)
}

func QueryVID(vid string) {
	client := pb.NewBackendClient(gRPCConn)
	ret, err := client.GetVidInfo(context.TODO(), &pb.GetVidInfoRequest{Vid: vid})
	if err != nil {
		log.Printf("查询失败：%s", err.Error())
		return
	}
	log.Printf("vid=%s, %s", vid, ret.GetMessage())
}

func RunUploadProcess(filePath string) {
	log.Printf("上传中……文件路径为：%s", filePath)
	vid, err := upload(filePath)
	if err != nil {
		log.Printf("上传失败：%s", err.Error())
		return
	}
	log.Printf("上传成功，vid=%s，转码中……", vid)

	client := pb.NewBackendClient(gRPCConn)
	stream, err := client.WatchUpload(context.Background(), &pb.WatchUploadRequest{Vid: vid})
	if err != nil {
		log.Printf("watch upload error: %s", err.Error())
		return
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("grpc stream error: %s", err.Error())
			return
		}
		log.Printf("vid=%s, %s", vid, msg.GetMessage())
	}
}

func upload(filePath string) (string, error) {
	uploadURL := fmt.Sprintf("http://%s:8081/upload", serverHost)
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open %s error: %s", filePath, err.Error())
	}
	defer f.Close()

	req, err := http.NewRequest(http.MethodPost, uploadURL, f)
	if err != nil {
		panic(err)
	}

	resp, err := httpCli.Do(req)
	if err != nil {
		return "", fmt.Errorf("httpCli do error: %s", err.Error())
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status=%s", resp.Status)
	}
	v, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body error: %s", err.Error())
	}
	return string(v), nil
}
