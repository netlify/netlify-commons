package pprof

import (
	"net/http"
	"net/http/pprof"

	log "github.com/sirupsen/logrus"
)

var ListenAddress = "localhost:6060"

// This init is needed to disable the default handlers registered during the init() from importing net/http/pprof
func init() {
	http.DefaultServeMux = http.NewServeMux()
}

// Handle adds the standard pprof handlers to the http.ServeMux
func Handle(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// NewServeMux builds a ServeMux and sets up the standard pprof handlers
func NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	Handle(mux)
	return mux
}

// NewServer constructs a new http.Server at addr with the standard pprof Mux
func NewServer(addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: NewServeMux(),
	}
}

// ListenAndServe starts a server at addr with standard pprof handlers.
func ListenAndServe(addr string) error {
	return NewServer(addr).ListenAndServe()
}

// Run a standard pprof server at ListenAddress
func Run() {
	go func() {
		log.Infof("Running pprof server at: %s", ListenAddress)
		log.Fatal(ListenAndServe(ListenAddress))
	}()
}
