package grpcapi

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

// Server struct for receiver methods
type server struct {
}

// Regfunc is a function to register custom server with
type regfunc func(server *grpc.Server)

// RegfuncREST is a function to register custom REST server with
type regfuncREST func(ctx context.Context, mux *runtime.ServeMux, opts []grpc.DialOption) error

// ServeGRPC ...
func ServeGRPC(bind string, registerFunc regfunc) {
	lis, err := net.Listen("tcp", bind)
	panicErr(err)
	s := grpc.NewServer()

	registerFunc(s)

	s.Serve(lis)
}

// ServeREST ...
func ServeREST(bind string, registerFunc regfuncREST, swaggerLocation string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		sw, err := ioutil.ReadFile(swaggerLocation)
		panicErr(err)
		io.Copy(w, bytes.NewReader(sw))

	})
	opts := []grpc.DialOption{grpc.WithInsecure()}

	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}))

	err := registerFunc(ctx, gwmux, opts)
	panicErr(err)
	mux.Handle("/", gwmux)

	http.ListenAndServe(bind, allowCORS(mux))
	panicErr(err)
}

// TODO: Netjes maken dit
// allowCORS allows Cross Origin Resoruce Sharing from any origin.
// Don't do this without consideration in production systems.
func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r)
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
func preflightHandler(w http.ResponseWriter, r *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
	glog.Infof("preflight request for %s", r.URL.Path)
	return
}
func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
