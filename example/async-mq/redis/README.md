# mq

MQ serves as proxy between "message queue" and remote receiver. It polls queue periodically, crafts single POST request from multiple messages and sends it to HTTP endpoint, exposed by receiver.

queue - interfaces to interact with queue (contains REDIS implementation)

log - interface for logger

server.go - HTTP server

proxy.go - implementation of proxy's logic
