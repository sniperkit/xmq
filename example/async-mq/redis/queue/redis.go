package queue

import(
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/redis.v3"
)

const (
	redisPipedCommandSize int = 200
	redisPipedCommandBufferSize int = 1000
)


type RedisConsumerConfig struct {
	Id       string `json:"id"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Db       int    `json:"db"`
	Queue    string `json:"queue"`
}

// REDIS implementation of Consumer
type RedisConsumer struct {
	id             string
	client         *redis.Client
	keys           *redisKeys
	chPipedCmds    chan redisPipedCommand
	chPipedFlushed chan struct{}
}

func NewRedisConsumer(config *RedisConsumerConfig) (*RedisConsumer, error) {
	// Create Redis client
	options := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       int64(config.Db),
	}
	client := redis.NewClient(&options)

	// Check if queue exists
	queueFound, err := client.Exists(config.Queue).Result()
	if err != nil {
		client.Close()
		return nil, err
	}
	if !queueFound {
		client.Close()
		return nil, fmt.Errorf("Queue %s does not exist", config.Queue)
	}

	keys := newReidsKeys(config.Queue, config.Id)

	// Ensure message ID counter
	err = client.SetNX(keys.msgCounter, -1, 0).Err()
	if err != nil {
		return nil, err
	}

	// Create consumer
	consumer := &RedisConsumer{
		id:             config.Id,
		client:         client,
		keys:           keys,
		chPipedCmds:    make(chan redisPipedCommand, redisPipedCommandBufferSize),
		chPipedFlushed: nil,
	}
	go consumer.processPipedCommands()

	return consumer, nil
}

func (rc *RedisConsumer) GetMessages(count int) ([]Message, error) {
	// Wait for first message, then fetch as many as possible
	cmders, err := rc.client.Pipelined(func(pipe *redis.Pipeline) error {
		pipe.BRPopLPush(rc.keys.queue, rc.keys.unacked, 0)
		for i := 1; i < count; i++ {
			pipe.RPopLPush(rc.keys.queue, rc.keys.unacked)
		}
		return nil
	})
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Convert Redis reply strings to messages
	msgs, err := rc.cmdersToMessages(cmders)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (rc *RedisConsumer) GetStats() (*ConsumerStats, error) {
	queue, err := rc.client.LLen(rc.keys.queue).Result()
	if err != nil {
		return nil, err
	}

	unacked, err := rc.client.LLen(rc.keys.unacked).Result()
	if err != nil {
		return nil, err
	}

	rejected, err := rc.client.LLen(rc.keys.rejected).Result()
	if err != nil {
		return nil, err
	}

	stats := &ConsumerStats{
		Queue:    int(queue),
		Unacked:  int(unacked),
		Rejected: int(rejected),
	}

	return stats, nil
}

func (rc *RedisConsumer) Close() error {
	close(rc.chPipedCmds)
	<- rc.chPipedFlushed
	return rc.Close()
}

func (rc *RedisConsumer) processPipedCommands() {
	cmds := make([]redisPipedCommand, 0, redisPipedCommandSize)

	// Collect predefined number of commands, then execute them in pipeline mode
	for cmd := range rc.chPipedCmds {
		cmds = append (cmds, cmd)
		if len(cmds) == redisPipedCommandSize {
			rc.client.Pipelined(func(pipe *redis.Pipeline) error {
				for i, _ := range cmds {
					cmds[i].execute(pipe)
					cmds[i] = nil
				}
				return nil
			})
			cmds = make([]redisPipedCommand, 0, redisPipedCommandSize)
		}
	}

	close(rc.chPipedFlushed)
}

func (rc *RedisConsumer) cmdersToMessages(cmders []redis.Cmder) ([]Message, error) {
	msgCounter, err := rc.client.Get(rc.keys.msgCounter).Int64()
	if err != nil {
		return nil, err
	}

	msgs := make([]Message, 0, len(cmders))
	for _, cmder := range cmders {
		switch cmd := cmder.(type) {
		case *redis.StringCmd:
			// Convert Redis string reply to message
			if cmd.Val() == "" {
				continue
			}
			msg, err := rc.stringCmdToMessage(cmd)
			if err != nil {
				return nil, err
			}
			// Craft message's id
			msgCounter++
			msg.payload.Id = fmt.Sprintf("%s:%d", rc.id, msgCounter)
			msgs = append(msgs, msg)
		default:
			return nil, errors.New("Invalid redis.Cmder type")
		}
	}

	rc.client.Set(rc.keys.msgCounter, msgCounter, 0)
	return msgs, nil
}

func (rc *RedisConsumer) stringCmdToMessage(strCmd *redis.StringCmd) (*redisMessage, error) {
	rawBody, err := strCmd.Bytes()
	if err != nil {
		return nil, err
	}

	jsonBody := make(map[string]interface{})
	err = json.Unmarshal(rawBody, jsonBody)
	if err != nil {
		return nil, err
	}

	hash := md5.Sum(rawBody)

	payload := &MessagePayload{
		Hash: string(hash[:]),
		Body: jsonBody,
	}
	msg := &redisMessage{
		rc,
		payload,
	}

	return msg, nil
}

// REDIS implementation of Message
type redisMessage struct {
	consumer *RedisConsumer
	payload  *MessagePayload
}

func (rm *redisMessage) Acknowledge() error {
	cmd := &redisPipedRPop{
		key: rm.consumer.keys.unacked,
	}
	rm.consumer.chPipedCmds <- cmd

	return nil
}

func (rm *redisMessage) Reject() error {
	cmd := &redisPipedRPopLPush{
		keySrc:  rm.consumer.keys.unacked,
		keyDest: rm.consumer.keys.rejected,
	}
	rm.consumer.chPipedCmds <- cmd

	return nil
}

func (rm *redisMessage) Payload() *MessagePayload {
	return rm.payload
}

// REDIS string keys in usage
type redisKeys struct {
	queue      string
	unacked    string
	rejected   string
	msgCounter string
}

func newReidsKeys(queue, consumer string) *redisKeys {
	return &redisKeys{
		queue:      queue,
		unacked:    fmt.Sprintf("%s::%s::unacked", queue, consumer),
		rejected:   fmt.Sprintf("%s::%s::rejected", queue, consumer),
		msgCounter: fmt.Sprintf("%s::%s::msgCounter", queue, consumer),
	}
}

// REDIS command, that will be executed in pipeline mode
type redisPipedCommand interface {
	execute(pipe *redis.Pipeline) error
}

// Pipelined RPOP command
type redisPipedRPop struct {
	key string
}

func (rPop *redisPipedRPop) execute(pipe *redis.Pipeline) error {
	return pipe.RPop(rPop.key).Err()
}

// Pipelined RPOPLPUSH command
type redisPipedRPopLPush struct {
	keySrc  string
	keyDest string
}

func (rPopLPush *redisPipedRPopLPush) execute(pipe *redis.Pipeline) error {
	return pipe.RPopLPush(rPopLPush.keySrc, rPopLPush.keyDest).Err()
}