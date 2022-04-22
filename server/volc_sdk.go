package main

import (
	"github.com/bytedance/sonic"
	"github.com/volcengine/volc-sdk-golang/base"
	"github.com/volcengine/volc-sdk-golang/service/vod"
	"github.com/volcengine/volc-sdk-golang/service/vod/models/request"
	"github.com/volcengine/volc-sdk-golang/service/vod/upload/functions"
	"log"
	"strings"
	"sync"
)

type VolcSDK struct {
	instance *vod.Vod
}

func NewVolcSDK(ak, sk string) *VolcSDK {
	instance := vod.NewInstanceWithRegion(base.RegionCnNorth1)
	instance.SetCredential(base.Credentials{
		AccessKeyID:     ak,
		SecretAccessKey: sk,
	})
	return &VolcSDK{
		instance: instance,
	}
}

var (
	volcOnce sync.Once
	volcSDK  *VolcSDK
)

func GetVolcSDK() *VolcSDK {
	volcOnce.Do(func() {
		volcSDK = NewVolcSDK("AKLTYTlkZTZkMDEzMTllNDQzODkzNDNiODE5OGI0Mzk3ZTk",
			"TUdJMU9XRTNOMlExWldSaE5HRmpOemxrTnpReU5EUTVOREF6WTJObE9UZw==")
	})

	return volcSDK
}

func (v *VolcSDK) GetVideoStatus(vidSet map[string]struct{}) (map[string]*VideoStatus, error) {
	vidList := make([]string, 0, len(vidSet))
	for k := range vidSet {
		vidList = append(vidList, k)
	}

	result := map[string]*VideoStatus{}

	for i := 0; i < len(vidList); i += 20 {
		size := len(vidList) - i
		if size > 20 {
			size = 20
		}

		vids := strings.Join(vidList[i:i+size], ",")
		query := &request.VodGetMediaInfosRequest{
			Vids: vids,
		}
		resp, _, err := v.instance.GetMediaInfos(query)
		if err != nil {
			log.Printf("GetMediaInfos error: %s", err.Error())
			continue
		}

		for _, mediaInfo := range resp.GetResult().MediaInfoList {
			published := false
			if mediaInfo.GetBasicInfo().GetPublishStatus() == "Published" {
				published = true
			}
			result[mediaInfo.GetBasicInfo().GetVid()] = &VideoStatus{Published: published}
		}
	}

	return result, nil
}

func (v *VolcSDK) UploadFile(filePath string) (vid string, err error) {
	spaceName := "shennong_zbw"

	snapShotFunc := functions.SnapshotFunc(1.3)
	getMetaFunc := functions.GetMetaFunc()
	workflowFunc := functions.StartWorkflowFunc("06853553c4d3402698a17ff5dff87fd7")
	funcList := []*vod.Function{
		&snapShotFunc,
		&getMetaFunc,
		&workflowFunc,
	}

	funcs, _ := sonic.MarshalString(funcList)

	resp, _, err := v.instance.UploadMediaWithCallback(&request.VodUploadMediaRequest{
		SpaceName:    spaceName,
		FilePath:     filePath,
		CallbackArgs: "my callback",
		Functions:    funcs,
	})

	if err != nil {
		return
	}
	return resp.GetResult().GetData().GetVid(), nil
}
