package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"gitlab.com/joukehofman/OTSthingy/proto"
	"google.golang.org/grpc"
)

var abort bool
var logs logger
var cfg *config

// our main func starts the GRPC API and REST grpc-gateway
// as well as a poller routine that checks if timestamp requests have been finalized
func main() {

	// init stuff.
	logs = logger{}
	logs.init(os.Stdout, os.Stdout, os.Stderr, os.Stdout)
	cfg = &config{}
	cfg.fromEnv()
	cfg.fromFlags() // flags overwrite environment vars

	_requester := requester{
		pendingRequests: make(map[string]*request),
		mutex:           &sync.Mutex{},
	}
	_notifier := notifier{}

	// functions to start grpc server with
	rfunc := func(server *grpc.Server) {
		OTSthingy.RegisterTimestampServer(server, timestampServer{
			requestr: &_requester,
		})
	}
	restfunc := func(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
		return OTSthingy.RegisterTimestampHandlerFromEndpoint(ctx, mux, cfg.grpcBind, opts)
	}

	// start grpc and rest API
	go serveGRPC(cfg.grpcBind, rfunc)
	go serveREST(cfg.restBind, restfunc, "api/api.swagger.json")

	// start poller.
	abortChan := make(chan bool, 5)
	notifyChan := make(chan *request, 5)

	poller := poller{
		interval:   100, // ms
		abortChan:  abortChan,
		notifyChan: notifyChan,
		_requester: &_requester,
		_notifier:  &_notifier,
	}

	for i := 0; i < cfg.notifiers; i++ {
		go poller.notify()
	}
	go poller.start()

	// wait for OS signal to shut down properly
	interuptChan := make(chan os.Signal)
	signal.Notify(interuptChan, os.Interrupt)
	s := <-interuptChan

	abort = true
	logs.debug.Println("Got signal:", s)
	logs.debug.Println("Waiting for poller and notifiers to finish...")
	go delayedForceExit()
	_ = <-abortChan // abort ack from poller

	// Send abort signal to notifiers
	for i := 0; i < cfg.notifiers; i++ {
		notifyChan <- nil
	}
	// Abort ack from notifiers
	for i := 0; i < cfg.notifiers; i++ {
		_ = <-abortChan
	}
	logs.debug.Println("Done")
}

func delayedForceExit() {
	time.Sleep(time.Second * 10)
	logs.errors.Println("Forcing exit after 10 secs")
	os.Exit(0)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
