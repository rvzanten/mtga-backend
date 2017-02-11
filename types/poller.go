package types

import "time"

// Poller polls for changes in timestamping status
type Poller struct {
	Abort      bool
	Url        string
	Interval   time.Duration // milliseconds
	AbortChan  chan bool
	NotifyChan chan *Request
	Requester  *Requester
	Notifier   *Notifier
}

func (poller *Poller) Start() {
	for !poller.Abort {
		poller.poll()
		time.Sleep(time.Millisecond * poller.Interval)
	}
	poller.AbortChan <- true
}

func (poller *Poller) poll() {
	req := poller.Requester
	req.Mutex.Lock()
	for hash, request := range req.PendingRequests {
		req.Mutex.Unlock()

		switch request.status {
		case STATUS_NEW:
			go request.process()
			break
		case STATUS_PENDING:
			// Do nothing for now
			// TODO: get info from tss API in separate request here
			break
		case STATUS_CONFIRMED:
			req.Mutex.Lock()
			delete(req.PendingRequests, hash)
			req.Mutex.Unlock()
			// TODO: update proof
			poller.NotifyChan <- request
			break
		}
		req.Mutex.Lock()
	}
	req.Mutex.Unlock()
}

func (poller *Poller) Notify() {
	for !poller.Abort {
		req := <-poller.NotifyChan
		logs.Debug.Println("Got request from chan")
		if req != nil {
			var emailErr, webhookErr error

			// TODO: extract txid
			txid := []byte{}
			if req.tsRequest.EmailAddress != "" {
				emailErr = poller.Notifier.sendMail(req, txid)
				if emailErr != nil {
					logs.Errors.Println(emailErr)
				}
			}
			if req.tsRequest.WebhookUrl != "" {
				webhookErr = poller.Notifier.callWebhook(req, txid)
				if webhookErr != nil {
					logs.Errors.Println(webhookErr)
				}
			}
		}
	}
	poller.AbortChan <- true
}
