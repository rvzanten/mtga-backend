package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"time"

	"gitlab.com/joukehofman/OTSthingy/proto"
)

type poller struct {
	url        string
	interval   time.Duration // milliseconds
	abortChan  chan bool
	notifyChan chan *request
	requestr   *requester
}

func (poller *poller) start() {
	for !abort {
		poller.poll()
		time.Sleep(time.Millisecond * poller.interval)
	}
	poller.abortChan <- true
}

func (poller *poller) poll() {
	req := poller.requestr
	req.mutex.Lock()
	for hash, request := range req.pendingRequests {
		req.mutex.Unlock()

		// TODO: get info from tss API
		isNowComplete := false

		if isNowComplete {
			req.mutex.Lock()
			delete(req.pendingRequests, hash)
			req.mutex.Unlock()
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
			if req.tsRequest.EmailAddress != "" {
				// TODO: extract txid
				txid := []byte{}
				poller.sendMail(req.tsRequest, txid)
			}
			if req.tsRequest.WebhookUrl != "" {
				// TODO: call webhook

			}
		}
	}
	poller.abortChan <- true
}

func (poller *poller) sendMail(tsReq *OTSthingy.TimeStampRequest, txid []byte) error {

	// auth := smtp.PlainAuth("", cfg.fromAddr, "password", cfg.smtpHost)
	c, err := smtp.Dial(cfg.smtpHost)
	if err != nil {
		return err
	}
	defer c.Close()
	c.Mail(cfg.fromAddr)
	c.Rcpt(tsReq.EmailAddress)
	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()
	buf := bytes.NewBufferString(fmt.Sprintf(
		"Subject: %s\r\n"+
			"\r\n"+
			"We would like to notify that your timestamp request (labelled '%s') has been finalized. The transaction hash:\r\n"+
			"\r\n"+
			"%s",
		"timestamp request adopted into Bitcoin blockchain",
		"label", //tsReq.Label,
		hex.EncodeToString(txid),
	))
	_, err = buf.WriteTo(wc)

	return err
}
