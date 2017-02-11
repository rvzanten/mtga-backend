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

		switch request.status {
		case STATUS_NEW:
			go request.process()
			break
		case STATUS_PENDING:
			// Do nothing for now
			break
		case STATUS_CONFIRMED:
			req.mutex.Lock()
			delete(req.pendingRequests, hash)
			req.mutex.Unlock()
			// TODO: update proof
			poller.notifyChan <- request
			break
		}
		req.mutex.Lock()
	}
	req.mutex.Unlock()
}

func (poller *poller) notify() {
	for !abort {
		req := <-poller.notifyChan
		logs.debug.Println("Got request from chan")
		if req != nil {
			var emailErr, webhookErr error

			// TODO: extract txid
			txid := []byte{}
			if req.tsRequest.EmailAddress != "" {
				emailErr = poller._notifier.sendMail(req, txid)
				if emailErr != nil {
					logs.errors.Println(emailErr)
				}
			}
			if req.tsRequest.WebhookUrl != "" {
				webhookErr = poller._notifier.callWebhook(req, txid)
				if webhookErr != nil {
					logs.errors.Println(webhookErr)
				}
			}

			// TODO: smart error handling
		}
	}
	poller.abortChan <- true
}
