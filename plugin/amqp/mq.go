// Package mq wrap the github.com/streadway/amqp library to be able to mock it.
package amqp

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// MessageQueue struct is compliant to the MessageQueue interface.
type MessageQueue struct {
	url        string
	name       string
	Reconnect  bool
	connection *amqp.Connection
	Channel    *amqp.Channel
	queue      amqp.Queue
}

// NewMQ open a connection to AMQP server, open a channel of communication and then return a MQ struct holding the connection and the channel.
func NewPersistent(amqpURL string) (mq MessageQueue, err error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return
	}
	ch, err := conn.Channel()
	if err != nil {
		return
	}
	mq = MessageQueue{
		connection: conn,
		Channel:    ch,
	}
	return
}

// New instantiate a MQ MessageQueue holding the connection and the channel.
func New(url string, name string) *MessageQueue {
	mq := new(MessageQueue)
	mq.SetUrl(url) // rabbitmq server
	mq.name = name
	mq.Reconnect = true
	return mq
}

func (mq *MessageQueue) SendMatch(match []string) {
	data, _ := json.Marshal(match)
	err := mq.Channel.Publish(
		"",
		mq.name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         data,
		})
	FailOnError(err, "Unable to publish match")
}

func (mq *MessageQueue) Connect() bool {
	var err error
	// start with 2 seconds before retry
	var stepping = 2
	// try connection until it succeeds
	mq.connection, err = amqp.Dial(mq.url)
	for err != nil {
		log.Printf("Unable to connect to RabbitMQ.  Retrying in %d seconds...\n", stepping)
		log.Printf(err.Error())
		// wait a bit
		time.Sleep(time.Duration(stepping) * time.Second)
		// increase time between attempts
		if stepping < 60 {
			stepping = stepping * 2
		}
		if !mq.Reconnect {
			return false
		}
		mq.connection, err = amqp.Dial(mq.url)
	}
	log.Printf("Connected to RabbitMQ")
	mq.Channel, err = mq.connection.Channel()
	FailOnError(err, "Failed to open a channel")

	mq.CreateQueues()

	return true
}

func (mq *MessageQueue) NoReconnect() {
	mq.Reconnect = false
}

func (mq *MessageQueue) SetUrl(url string) {
	mq.url = url
}

func (mq *MessageQueue) CreateQueues() {
	var err error
	log.Printf("Declaring queue '%s'", mq.name)
	mq.queue, err = mq.Channel.QueueDeclare(
		mq.name, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	FailOnError(err, "Failed to declare recv queue")
}

// DeclareQueue declare a queue an set QoS
func (mq *MessageQueue) DeclareQueue(queueName string) error {
	_, err := mq.Channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return fmt.Errorf("Failed to declare a queue: %v", err)
	}

	return mq.Channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
}

// Publish provide a way to publish a message containing
// the provided body to the queue with name queueName
func (mq *MessageQueue) Publish(queueName string, body []byte) error {
	return mq.Channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
}

// Consume will start listening to the message queue using the provided queue name.
// It will call the Receiver function every time a message arrives.
func (mq *MessageQueue) Consume(queueName string, r Receiver) error {
	msgs, err := mq.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			msg := Message{
				delivery: d,
			}
			r(msg, forever)
		}
		forever <- true
	}()
	<-forever
	return nil
}

func (mq *MessageQueue) GetSubscriptionChannel() <-chan amqp.Delivery {
	msgs, err := mq.Channel.Consume(
		mq.queue.Name, // queue
		"",            // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	FailOnError(err, "Failed to register a consumer")
	return msgs
}

func (mq *MessageQueue) Close() {
	mq.Channel.Close()
	mq.connection.Close()
}

// Message is the wrapper for the Delivery struct of the github.com/streadway/amqp library.
// Its purpose is to met the be compliant with the Delivery interface in this package (mq)
// and help making the rest of the project testable.
type Message struct {
	delivery amqp.Delivery
}

// Body will return the body of the message.
func (m Message) Body() []byte {
	return m.delivery.Body
}

// Ack delivers an acknowledgment that the message has been receive and treated.
// The multiple argument is true when the all the previous messages can be acknowledged as well.
func (m Message) Ack(multiple bool) error {
	return nil
}

// Nack delivers a negative acknowledgment signifying a failure in treating the message.
// If multiple is true, all the previous messages that weren't aknowledged yet are going
// to be negatively aknowledged.
// If requeue is true, it means that the message needs to be requeued.
func (m Message) Nack(multiple, requeue bool) error {
	return nil
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
