package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	VidStateUploading = iota
	VidStateTranscoding
	VidStateUnPublished
	VidStatePublished
)

var (
	VidStateChineseMap = map[uint8]string{
		VidStateUploading:   "上传中",
		VidStateTranscoding: "转码中",
		VidStateUnPublished: "未发布",
		VidStatePublished:   "已发布",
	}
)

type VidStatus struct {
	LastUpdate time.Time
	Vid        string
	State      uint8
}

type DAL interface {
	InTranscoding(vid string) error               // state: uploading -> transcoding
	FinishTranscoding(vid string) error           // state: transcoding -> unpublished
	Publish(vid string, info interface{}) error   // only changed on UnPublished state, state: unpublished -> published
	UnPublish(vid string, info interface{}) error // only changed on Published state, state: published -> unpublished
	Get(vid string) (*VidStatus, error)
}

func NewDALInMem() DAL {
	return &DALInMemImpl{
		data: make(map[string]*VidStatus),
	}
}

type DALInMemImpl struct {
	sync.RWMutex
	data map[string]*VidStatus
}

var (
	ErrOutdatedRequest = fmt.Errorf("outdated request")
	ErrVidNotFound     = fmt.Errorf("not found")
)

func (d *DALInMemImpl) InTranscoding(vid string) error {
	var vidStatus *VidStatus
	now := time.Now()

	d.Lock()
	defer d.Unlock()
	if v, ok := d.data[vid]; ok {
		vidStatus = v
		if now.Before(vidStatus.LastUpdate) {
			return ErrOutdatedRequest
		}
	} else {
		vidStatus = &VidStatus{
			Vid:   vid,
			State: VidStateTranscoding,
		}
	}

	vidStatus.State = VidStateTranscoding
	vidStatus.LastUpdate = now
	d.data[vid] = vidStatus
	return nil
}

func (d *DALInMemImpl) FinishTranscoding(vid string) error {
	now := time.Now()

	d.Lock()
	defer d.Unlock()

	v, ok := d.data[vid]
	if !ok {
		return ErrVidNotFound
	}
	if now.Before(v.LastUpdate) {
		return ErrOutdatedRequest
	}
	if v.State == VidStateTranscoding {
		v.State = VidStateUnPublished
		v.LastUpdate = now
	}
	return nil
}

func (d *DALInMemImpl) Publish(vid string, info interface{}) error {
	now := time.Now()

	d.Lock()
	defer d.Unlock()

	v, ok := d.data[vid]
	if !ok {
		return ErrVidNotFound
	}
	if now.Before(v.LastUpdate) {
		return ErrOutdatedRequest
	}
	if v.State == VidStateUnPublished {
		v.State = VidStatePublished
		v.LastUpdate = now
	}
	return nil
}

func (d *DALInMemImpl) UnPublish(vid string, info interface{}) error {
	now := time.Now()

	d.Lock()
	defer d.Unlock()

	v, ok := d.data[vid]
	if !ok {
		return ErrVidNotFound
	}
	if now.Before(v.LastUpdate) {
		return ErrOutdatedRequest
	}
	if v.State == VidStatePublished {
		v.State = VidStateUnPublished
		v.LastUpdate = now
	}
	return nil
}

func (d *DALInMemImpl) Get(vid string) (*VidStatus, error) {
	d.RLock()
	defer d.RUnlock()

	v, ok := d.data[vid]
	if !ok {
		log.Printf("query vid: %s, vids: %+v", vid, d.data)
		return nil, ErrVidNotFound
	}
	return v, nil
}

/*
v0dc8cg10001c9ft4nptn37ec3dh4hj0
v0dc8cg10001c9ft4nptn37ec3dh4hj0
*/
