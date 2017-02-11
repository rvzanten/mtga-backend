package main

import (
	"errors"
	"sync"

	"gitlab.com/joukehofman/OTSthingy/proto"
)

type request struct {
	incompleteTS []byte
	tsRequest    *OTSthingy.TimeStampRequest
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
	r.mutex.Unlock()

	// TODO: push request to API

	r.mutex.Lock()
	r.pendingRequests[string(tsReq.DocumentHash)] = &request{
		incompleteTS: []byte{},
		tsRequest:    tsReq,
	}
	r.mutex.Unlock()
	return nil
}
