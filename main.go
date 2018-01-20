package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mihis/mq/queue"
	"github.com/mihis/mq/log"
)

const (
	configFilePath string = "/config/config.json"
	logFilePath string = "/logs/request.log"
)


type serviceLogConfig struct {
	Enabled bool       `json:"enabled"`
	Format  log.Format `json:"format"`
	Level   log.Level  `json:"level"`
}

type serviceConfig struct {
	Server  mqServerConfig            `json:"server"`
	Proxy   mqProxyConfig             `json:"proxy"`
	Redis   queue.RedisConsumerConfig `json:"redis"`
	Log     serviceLogConfig          `json:"log"`
}


func main() {
	// Read configuration
	config, err := readConfigFile()
	if err != nil {
		fmt.Printf("Can not read configuration file; %s \n", err.Error())
		os.Exit(1)
	}

	// Create logger
	logger, err := createLogger(&config.Log)
	if err != nil {
		fmt.Printf("Can not create logger; %s \n", err.Error())
		os.Exit(1)
	}

	// Create message consumer
	consumer, err := queue.NewRedisConsumer(&config.Redis)
	if err != nil {
		fmt.Printf("Can not create REDIS message consumer; %s \n", err.Error())
		os.Exit(1)
	}

	// Create and start proxy service
	proxy, err := newProxy(consumer, logger, &config.Proxy)
	if err != nil {
		fmt.Printf("Can not create proxy service; %s \n", err.Error())
		os.Exit(1)
	}
	if err != proxy.start() {
		fmt.Printf("Can not start proxy service; %s \n", err.Error())
		os.Exit(1)
	}
	// Catch SIGTERM signal to terminate proxy service gracefully
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGTERM)
	go func(){
		<- chSignal
		proxy.stop()
	}()

	// Start HTTP server
	server, err := newServer(proxy, logger, &config.Server)
	if err != nil {
		fmt.Printf("Can not create HTTP server; %s \n", err.Error())
		os.Exit(1)
	}
	if err = server.start(); err != nil {
		fmt.Printf("Can not start HTTP server; %s \n", err.Error())
		os.Exit(1)
	}
}

func readConfigFile() (*serviceConfig, error) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	configRaw, err := ioutil.ReadFile(filepath.Join(dir,configFilePath))
	if err != nil {
		return nil, err
	}

	config := &serviceConfig{}
	if err := json.Unmarshal(configRaw, config); err != nil {
		return nil, err
	}

	return config, nil
}

func createLogger(config *serviceLogConfig) (log.Interface, error) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	var logFile io.Writer
	if config.Enabled {
		logFile, err := os.Create(filepath.Join(dir, logFilePath))
		if err != nil {
			return nil, err
		}
		logFile = logFile
	} else {
		logFile = ioutil.Discard
	}

	log, err := log.NewLogrus(logFile, config.Level, config.Format)
	if err != nil {
		return nil, err
	}

	return log, nil
}