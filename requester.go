package main

import (
	"errors"
	"sync"
	"time"

	"gitlab.com/joukehofman/OTSthingy/proto"
)

type requestStatus int

const (
	STATUS_NEW requestStatus = iota
	STATUS_PENDING
	STATUS_CONFIRMED
)

type request struct {
	proof     []byte
	status    requestStatus
	tsRequest *OTSthingy.TimeStampRequest
}
type requester struct {
	pendingRequests map[string]*request
	mutex           *sync.Mutex
}

func (r *requester) addRequest(tsReq *OTSthingy.TimeStampRequest) error {

	r.mutex.Lock()
	if _, exists := r.pendingRequests[string(tsReq.DocumentHash)]; exists {
		r.mutex.Unlock()
		return errors.New("Request already exists")
	}

	// TODO: push the request to timestamp server here, and poll for status change separately
	// for now, we will add the request with status new so the poller will pick it up
	r.pendingRequests[string(tsReq.DocumentHash)] = &request{
		proof:     []byte{},
		tsRequest: tsReq,
		status:    STATUS_NEW,
	}
	r.mutex.Unlock()
	return nil
}

func (r *request) process() {
	r.status = STATUS_PENDING

	// TODO call to timestamp server script
	// exec('script', r.tsRequest.DocumentHash)
	time.Sleep(time.Second * 5)
	r.proof = []byte("Dit is het bewijs")
	r.status = STATUS_CONFIRMED
}
