package main

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"gitlab.com/joukehofman/OTSthingy/grpcapi"
	"gitlab.com/joukehofman/OTSthingy/proto"
	"gitlab.com/joukehofman/OTSthingy/types"
	"google.golang.org/grpc"
)

var logs *types.Logger
var cfg *types.Config
var poller *types.Poller

// our main func starts the GRPC API and REST grpc-gateway
// as well as a poller routine that checks if timestamp requests have been finalized
func main() {
	initVars()
	_requester := types.Requester{
		PendingRequests: make(map[string]*types.Request),
		Mutex:           &sync.Mutex{},
	}
	_notifier := types.Notifier{}
	abortChan := make(chan bool, 5)
	notifyChan := make(chan *types.Request, 5)

	startPoller(&_requester, &_notifier, abortChan, notifyChan)
	startGRPC(&_requester)

	// wait for OS signal to shut down properly
	interuptChan := make(chan os.Signal)
	signal.Notify(interuptChan, os.Interrupt)
	s := <-interuptChan

	poller.Abort = true
	logs.Debug.Println("Got signal:", s)
	logs.Debug.Println("Waiting for poller and notifiers to finish...")
	go delayedForceExit()
	_ = <-abortChan // abort ack from poller

	// Send abort signal to notifiers
	for i := 0; i < cfg.Notifiers; i++ {
		notifyChan <- nil
	}
	// Abort ack from notifiers
	for i := 0; i < cfg.Notifiers; i++ {
		_ = <-abortChan
	}
	logs.Debug.Println("Done")
}

func startPoller(requester *types.Requester, notifier *types.Notifier, abortChan chan bool, notifyChan chan *types.Request) {
	poller = &types.Poller{
		Interval:   100, // ms
		AbortChan:  abortChan,
		NotifyChan: notifyChan,
		Requester:  requester,
		Notifier:   notifier,
	}

	for i := 0; i < cfg.Notifiers; i++ {
		go poller.Notify()
	}
	go poller.Start()
}

func startGRPC(requester *types.Requester) {
	// functions to start grpc server with
	rfunc := func(server *grpc.Server) {
		OTSthingy.RegisterTimestampServer(server, grpcapi.TimestampServer{
			Requester: requester,
		})
	}
	restfunc := func(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
		return OTSthingy.RegisterTimestampHandlerFromEndpoint(ctx, mux, cfg.GrpcBind, opts)
	}

	// start grpc and rest API
	go grpcapi.ServeGRPC(cfg.GrpcBind, rfunc)
	go grpcapi.ServeREST(cfg.RestBind, restfunc, "api/api.swagger.json")
}

func initVars() {
	// init stuff.
	logs = &types.Logger{}
	logs.Init(os.Stdout, os.Stdout, os.Stderr, os.Stdout)
	cfg = &types.Config{}
	types.InitVars(logs, cfg)

	cfg.FromFlags()
	cfg.FromEnv() // ENV vars overwrite flags because of default values
}

func delayedForceExit() {
	time.Sleep(time.Second * 10)
	logs.Errors.Println("Forcing exit after 10 secs")
	os.Exit(0)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
