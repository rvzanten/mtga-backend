package main

import (
	"flag"
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
var requestr requester
var logs logger

// our main func starts the GRPC API and REST grpc-gateway
// as well as a poller routine that checks if timestamp requests have been finalized
func main() {

	// init stuff.
	logs = logger{}
	logs.init(os.Stdout, os.Stdout, os.Stderr, os.Stdout)
	var (
		grpcBind  = flag.String("grpcBind", ":8181", "Expose storeapi GRPC on this port")
		restBind  = flag.String("restBind", ":8080", "Expose storeapi REST api on this port")
		tssURL    = flag.String("tssURL", "", "The API URL of the timestamp server")
		notifiers = flag.Int("notifiers", 5, "Amount of notifier routines to run")
	)
	flag.Parse()
	requestr = requester{
		url:             *tssURL,
		pendingRequests: make(map[string]*request),
		mutex:           &sync.Mutex{},
	}

	// functions to start grpc server with
	rfunc := func(server *grpc.Server) {
		OTSthingy.RegisterTimestampServer(server, timestampServer{})
	}
	restfunc := func(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error {
		return OTSthingy.RegisterTimestampHandlerFromEndpoint(ctx, mux, *grpcBind, opts)
	}

	// start grpc and rest API
	go serveGRPC(*grpcBind, rfunc)
	go serveREST(*restBind, restfunc, "api/api.swagger.json")

	// start poller.
	abortChan := make(chan bool, 5)
	notifyChan := make(chan *request, 5)

	poller := poller{
		url:        *tssURL,
		interval:   100,
		abortChan:  abortChan,
		notifyChan: notifyChan,
	}

	for i := 0; i < *notifiers; i++ {
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
	for i := 0; i < *notifiers; i++ {
		notifyChan <- nil
	}
	// Abort ack from notifiers
	for i := 0; i < *notifiers; i++ {
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
