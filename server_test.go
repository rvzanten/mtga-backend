package main

import (
	"testing"

	"gitlab.com/joukehofman/OTSthingy/types"
)

func TestInitVars(t *testing.T) {
	initVars()
	if logs == nil {
		t.Fail()
	}
}

func TestStartGRPC(t *testing.T) {
	startGRPC(&types.Requester{})
}

func TestStartPoller(t *testing.T) {
	abortChan := make(chan bool, 5)
	notifyChan := make(chan *types.Request, 5)
	startPoller(&types.Requester{}, &types.Notifier{}, abortChan, notifyChan)
	go poller.Notify()
	poller.Abort = true

	_ = <-abortChan
}
