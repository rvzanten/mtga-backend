package main

import "time"

type poller struct {
	url        string
	interval   time.Duration // milliseconds
	abortChan  chan bool
	notifyChan chan *request
}

func (poller *poller) start() {
	for !abort {
		poller.poll()
		time.Sleep(time.Millisecond * poller.interval)
	}
	poller.abortChan <- true
}

func (poller *poller) poll() {
	requestr.mutex.Lock()
	for hash, request := range requestr.pendingRequests {
		requestr.mutex.Unlock()

		// TODO: get info from tss API
		isNowComplete := false

		if isNowComplete {
			requestr.mutex.Lock()
			delete(requestr.pendingRequests, hash)
			requestr.mutex.Unlock()
			poller.notifyChan <- request
		}
		requestr.mutex.Lock()
	}
	requestr.mutex.Unlock()
}

func (poller *poller) notify() {
	for !abort {
		req := <-poller.notifyChan
		if req != nil {
			if req.email != "" {
				// TODO: send email
			}
			if req.webhook != "" {
				// TODO: call webhook
			}
		}
	}
	poller.abortChan <- true
}
