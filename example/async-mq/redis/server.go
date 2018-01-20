package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/mihis/mq/log"
)

// MQ service HTTP server configuration
type mqServerConfig struct {
	Port      int               `json:"port"`
	Endpoints mqServerEndpoints `json:"endpoints"`
}

type mqServerEndpoints struct {
	Stats string `json:"stats"`
}

// MQ service HTTP server
type mqServer struct {
	proxy     *mqProxy
	log       log.Interface
	port      int
	endpoints *mqServerEndpoints
}

func newServer(proxy *mqProxy, log log.Interface, config *mqServerConfig) (*mqServer, error) {
	server := &mqServer{
		proxy:     proxy,
		log:       log,
		port:      config.Port,
		endpoints: &config.Endpoints,
	}

	return server, nil
}

func (server *mqServer) start() error {
	router := httprouter.New()
	router.Handle("GET", server.endpoints.Stats, server.wrapHanlder(server.geStats))

	addr := fmt.Sprintf("localhost:%s", server.port)
	return http.ListenAndServe(addr, router)
}

// Protects and logs HTTP handlers
func (server *mqServer) wrapHanlder(handler httprouter.Handle) httprouter.Handle {
	panicHanlder := func() {
		if err := recover(); err != nil {
			server.log.Error(log.Fields{"event": "panic", "reason": err})
		}
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		defer panicHanlder()
		handler(w, r, p)
	}
}

// Retrieves consumer's statistics
func (server *mqServer) geStats(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Get consumer's statistics
	stats, err := server.proxy.queue.GetStats()
	if err != nil {
		server.log.Error(log.Fields{"event": "HTTP handler failed", "reason": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert stats to JSON
	statsJson, err := json.Marshal(stats)
	if err != nil {
		server.log.Error(log.Fields{"event": "HTTP handler failed", "reason": err.Error()})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(statsJson)
}


