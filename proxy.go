package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/mihis/mq/queue"
	"github.com/mihis/mq/log"
)

// MQ service configuration
type mqProxyConfig struct {
	Name      string   `json:"name"`
	IpList    []string `json:"ipList"`
	PollCount int      `json:"pollCount"`
	PollRate  float64  `json:"pollRate"`
	Receiver  string   `json:"receiver"`
	Timeout   int      `json:"timeout"`
}

// MQ service
type mqProxy struct {
	queue      queue.Consumer
	log        log.Interface
	httpClient *http.Client
	name       string
	ipList     map[string]struct{}
	pollCount  int
	pollRate   time.Duration
	receiver   string
	chStop     chan chan error
}

func newProxy(queue queue.Consumer, log log.Interface, config *mqProxyConfig) (*mqProxy, error) {
	// Validate receiver's URL
	if _, err := url.Parse(config.Receiver); err != nil {
		return nil, fmt.Errorf("Invalid receiver URL; %s", err.Error())
	}

	proxy := &mqProxy{
		name:      config.Name,
		queue:     queue,
		log:       log,
		ipList:    nil,
		pollCount: config.PollCount,
		pollRate:  time.Millisecond * time.Duration(config.PollRate * 1000),
		receiver:  config.Receiver,
		chStop:    make(chan chan error),
	}

	// Configure HTTP transport
	transport := &http.Transport{ }
	proxy.httpClient = &http.Client{
		Transport: transport,
		Timeout:   time.Second * time.Duration(config.Timeout),
	}

	// Craft IP white list
	if config.IpList != nil {
		proxy.ipList = make(map[string]struct{})
		for _, ip := range config.IpList {
			proxy.ipList[ip] = struct{}{}
		}
	}

	return proxy, nil
}

// Start MQ service
func (proxy *mqProxy) start() error {
	panicHandler := func() {
		if err := recover(); err != nil {
			proxy.log.Error(log.Fields{"event": "panic", "reason": err})
		}
	}

	go func() {
		defer panicHandler()
		proxy.run()
	}()

	return nil
}

// Terminate MQ service gracefully
func (proxy *mqProxy) stop() error {
	resp := make(chan error)
	proxy.chStop <- resp
	return <- resp
}

// Poll messages from queue periodically
func (proxy *mqProxy) run() {
	for {
		select {
		case _ = <- time.After(proxy.pollRate):
			proxy.processMessages()
		case resp := <- proxy.chStop:
			resp <- proxy.queue.Close()
			return
		}
	}
}

// Handle messages, extracted from queue
func (proxy *mqProxy) processMessages() {
	// Poll message queue
	msgs, err := proxy.queue.GetMessages(proxy.pollCount)
	if err != nil {
		proxy.log.Error(log.Fields{"event": "message consumer failed", "reason": err.Error()})
		return
	}

	// Filter messages by IP
	accepted := make([]queue.Message, 0, len(msgs))
	for _, m := range msgs {
		ip := m.Payload().Body["ip"].(string)
		if proxy.checkIp(ip) {
			accepted = append(accepted, m)
		} else {
			proxy.rejectMessage(m)
		}
	}

	// Craft HTTP request
	reqBody, err := json.Marshal(accepted)
	if err != nil {
		proxy.log.Error(log.Fields{"event": "proxy request json encoding failed", "resaon": err.Error()})
		return
	}
	req, err := http.NewRequest("POST", proxy.receiver, bytes.NewReader(reqBody))
	if err != nil {
		proxy.log.Error(log.Fields{"event": "proxy request creation failed", "reason": err.Error()})
		return
	}

	// Send HTTP request to receiver
	resp, err := proxy.httpClient.Do(req)
	if err != nil {
		proxy.log.Error(log.Fields{"event": "proxy request failed", "reason": err.Error()})
		return
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	proxy.acknowledgeMessages(msgs, resp.StatusCode)

	return
}

// Validate IP
func (proxy *mqProxy) checkIp(ip string) bool {
	if proxy.ipList != nil {
		_, found := proxy.ipList[ip]
		return found
	} else {
		return true
	}
}

func (proxy *mqProxy) acknowledgeMessages(msgs []queue.Message, httpCode int) {
	for _, m := range msgs {
		if err := m.Acknowledge(); err == nil {
			proxy.log.Info(log.Fields{"event": " message processed", "httpCode:": httpCode, "body": m.Payload().Body})
		} else {
			proxy.log.Error(log.Fields{"event": "message acknowledge failed", "reason": err.Error(), "body": m.Payload().Body})
		}
	}
}

func (proxy *mqProxy) rejectMessage(msg queue.Message) {
	if err := msg.Reject(); err == nil {
		proxy.log.Info(log.Fields{"event": "message rejected", "body": msg.Payload().Body})
	} else {
		proxy.log.Error(log.Fields{"event": "message reject failed", "reason": err.Error(), "body": msg.Payload().Body})
	}
}