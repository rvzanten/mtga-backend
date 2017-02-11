package main

import "time"

type poller struct {
	url        string
	interval   time.Duration // milliseconds
	abortChan  chan bool
	notifyChan chan *request
	_requester *requester
	_notifier  *notifier
}

func (poller *poller) start() {
	for !abort {
		poller.poll()
		time.Sleep(time.Millisecond * poller.interval)
	}
	poller.abortChan <- true
}

func (poller *poller) poll() {
	req := poller._requester
	req.mutex.Lock()
	for hash, request := range req.pendingRequests {
		req.mutex.Unlock()

		// TODO: get info from tss API
		isNowComplete := false

		if isNowComplete {
			req.mutex.Lock()
			delete(req.pendingRequests, hash)
			req.mutex.Unlock()
			// TODO: update proof
			request.proof = []byte{}
			poller.notifyChan <- request
		}
		req.mutex.Lock()
	}
	req.mutex.Unlock()
}

func (poller *poller) notify() {
	for !abort {
		req := <-poller.notifyChan
		if req != nil {
			var emailErr, webhookErr error

			// TODO: extract txid
			txid := []byte{}
			if req.tsRequest.EmailAddress != "" {
				emailErr = poller._notifier.sendMail(req, txid)
			}
			if req.tsRequest.WebhookUrl != "" {
				webhookErr = poller._notifier.callWebhook(req, txid)
			}

			// TODO: error handling
			if emailErr != nil && webhookErr != nil {
				// readd to queue?
			}
		}
	}
	poller.abortChan <- true
}
