package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"sync"

	"gitlab.com/joukehofman/OTSthingy/proto"
)

type requestStatus int

const (
	STATUS_NEW requestStatus = iota
	STATUS_PENDING
	STATUS_CONFIRMED
)

// Request defines a request and its status
type Request struct {
	proof     []byte
	status    requestStatus
	tsRequest *OTSthingy.TimeStampRequest
}

// Requester adds and processes timestamping requests
type Requester struct {
	PendingRequests map[string]*Request
	Mutex           *sync.Mutex
}

// AddRequest adds timestamp request
func (r *Requester) AddRequest(tsReq *OTSthingy.TimeStampRequest) error {

	r.Mutex.Lock()
	if _, exists := r.PendingRequests[string(tsReq.DocumentHash)]; exists {
		r.Mutex.Unlock()
		return errors.New("Request already exists")
	}

	// TODO: push the request to timestamp server here, and poll for status change separately
	// for now, we will add the request with status new so the poller will pick it up
	r.PendingRequests[string(tsReq.DocumentHash)] = &Request{
		proof:     []byte{},
		tsRequest: tsReq,
		status:    STATUS_NEW,
	}
	r.Mutex.Unlock()
	return nil
}

func (r *Request) process() {
	r.status = STATUS_PENDING
	docHex := hex.EncodeToString(r.tsRequest.DocumentHash)
	cmd := exec.Command("/home/jouke/GIT/gitlab.com/StrikerBee/opentimestamps-client-hash/ots", "-w", "stamp", "-hash", docHex, "-c", "http://10.20.100.70", "-m", "1")
	err := cmd.Run()
	if err != nil {
		logs.Errors.Printf("Was not able to execute stamping client: %s", err)
	} else {
		proofBytes, err := ioutil.ReadFile(fmt.Sprintf("%s.ots", docHex))
		if err != nil {
			logs.Errors.Printf("Could not read the proof file: %s, error: %s", fmt.Sprintf("%s.ots", docHex), err)
		}
		r.proof, r.status = proofBytes, STATUS_CONFIRMED
	}
}
