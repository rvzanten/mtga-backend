package main

import (
	"errors"
	"sync"
)

type request struct {
	incompleteTS []byte
	webhook      string
	email        string
}
type requester struct {
	url             string
	pendingRequests map[string]*request
	mutex           *sync.Mutex
}

func (r *requester) addRequest(hash []byte, webhook string, email string) error {

	r.mutex.Lock()
	if _, exists := r.pendingRequests[string(hash)]; exists {
		r.mutex.Unlock()
		return errors.New("Request already exists")
	}
	r.mutex.Unlock()

	// TODO: push request to API

	r.mutex.Lock()
	r.pendingRequests[string(hash)] = &request{
		incompleteTS: []byte{},
		webhook:      webhook,
		email:        email,
	}
	r.mutex.Unlock()
	return nil
}
