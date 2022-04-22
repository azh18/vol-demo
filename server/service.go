package main

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

type Watcher interface {
	ResultChan() <-chan string
	Close()
}

type Service interface {
	OnUploadSuccess(vid string, info interface{})
	OnTranscodeFinish(vid string, info interface{})
	Watch(vid string) (Watcher, error)
	GetVidInfo(vid string) (string, error)
	DAL() DAL
}

func NewService() Service {
	impl := &ServiceImpl{
		watchers: list.New(),
		db:       NewDALInMem(),
	}
	go impl.run()
	return impl
}

var (
	o         sync.Once
	myService Service
)

func GetService() Service {
	o.Do(func() {
		myService = NewService()
	})
	return myService
}

type watcherImpl struct {
	vid               string
	messageChan       chan string
	closed            bool
	closeMutex        sync.Mutex
	messageLastUpdate time.Time

	db DAL
}

func (w *watcherImpl) ResultChan() <-chan string {
	return w.messageChan
}

func (w *watcherImpl) Close() {
	w.closeMutex.Lock()
	defer w.closeMutex.Unlock()

	close(w.messageChan)
	w.closed = true
}

// get the latest status ant put into w.messageChan, with deduplication
func (w *watcherImpl) flushFromDAL() error {
	result, err := w.db.Get(w.vid)
	if err != nil {
		return err
	}
	w.closeMutex.Lock()
	defer w.closeMutex.Unlock()

	if !w.closed && result.LastUpdate.After(w.messageLastUpdate) {
		w.messageChan <- makeMessage(result)
		result.LastUpdate = w.messageLastUpdate
	}
	return nil
}

func makeMessage(status *VidStatus) string {
	return fmt.Sprintf("status=%s, lastUpdate=%s", VidStateChineseMap[status.State], status.LastUpdate.String())
}

type ServiceImpl struct {
	watchers *list.List

	db DAL
}

func (s *ServiceImpl) OnUploadSuccess(vid string, info interface{}) {
	if err := s.DAL().InTranscoding(vid); err != nil {
		log.Printf("call OnUploadSuccess on vid %s error: %s", vid, err.Error())
	}
}

func (s *ServiceImpl) OnTranscodeFinish(vid string, info interface{}) {
	if err := s.DAL().FinishTranscoding(vid); err != nil {
		log.Printf("call FinishTranscoding on vid %s error: %s", vid, err.Error())
	}
	s.triggerWatcherFlush(false, vid)
}

func (s *ServiceImpl) Watch(vid string) (Watcher, error) {
	w := &watcherImpl{
		vid:         vid,
		messageChan: make(chan string, 10000),
		db:          s.DAL(),
	}
	// make first message
	if err := w.flushFromDAL(); err != nil {
		return nil, err
	}
	s.watchers.PushBack(w)
	return w, nil
}

func (s *ServiceImpl) GetVidInfo(vid string) (string, error) {
	v, err := s.db.Get(vid)
	if err != nil {
		return "", err
	}
	return makeMessage(v), nil
}

func (s *ServiceImpl) DAL() DAL {
	return s.db
}

func (s *ServiceImpl) run() {
	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()
	for range ticker.C {
		s.refreshPublishStatus()
	}
}

func (s *ServiceImpl) triggerWatcherFlush(all bool, vid string) {
	for e := s.watchers.Front(); e != nil; e = e.Next() {
		w := e.Value.(*watcherImpl)
		if !all && vid != w.vid {
			continue
		}
		if err := w.flushFromDAL(); err != nil {
			log.Printf("flushFromDAL error: %s, vid=%s", err.Error(), w.vid)
		}
	}
}

func (s *ServiceImpl) refreshPublishStatus() {
	vidSet := map[string]struct{}{}
	var elemsToDelete []*list.Element
	for e := s.watchers.Front(); e != nil; e = e.Next() {
		w := e.Value.(*watcherImpl)
		if w.closed {
			elemsToDelete = append(elemsToDelete, e)
			continue
		}
		vidSet[w.vid] = struct{}{}
	}
	for _, e := range elemsToDelete {
		s.watchers.Remove(e)
	}

	// query vids
	videoStatusMap, err := s.GetVideoStatus(vidSet)
	if err != nil {
		return
	}

	for vid := range vidSet {
		videoStatus := videoStatusMap[vid]
		if videoStatus == nil {
			continue
		}
		if videoStatus.Published {
			if err := s.DAL().Publish(vid, nil); err != nil {
				log.Printf("call Publish on %s error: %s", vid, err.Error())
			}
		} else {
			if err := s.DAL().UnPublish(vid, nil); err != nil {
				log.Printf("call UnPublish on %s error: %s", vid, err.Error())
			}
		}
	}

	s.triggerWatcherFlush(true, "")
}

type VideoStatus struct {
	Published bool
}

func (s *ServiceImpl) GetVideoStatus(vidSet map[string]struct{}) (map[string]*VideoStatus, error) {
	return GetVolcSDK().GetVideoStatus(vidSet)
}
