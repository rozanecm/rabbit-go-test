package main

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/isayme/go-amqp-reconnect/rabbitmq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

var once sync.Once

type tRabbit struct {
	connection *rabbitmq.Connection
}

type IRabbit interface {
	ConsumeFromQueue(queueName, routingKey, exchangeName string, exclusive, exclusiveConsumer bool, name string) (<-chan amqp.Delivery, func(), error)
	CreateExchange(exchangeName string) error
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	CloseConnection()
}

var rabbit IRabbit

const (
	channelExhaustedError = "Exception (504) Reason: \"channel id space exhausted\""
)

func GetInstance() IRabbit {
	once.Do(func() {
		connection, err := rabbitmq.Dial(viper.GetString("rabbitmq.uri"))
		if err != nil {
			panic(err)
		}

		rabbit = &tRabbit{
			connection: connection,
		}
	})
	return rabbit
}

func (r *tRabbit) ConsumeFromQueue(queueName, routingKey, exchangeName string, exclusive, exclusiveConsumer bool, name string) (<-chan amqp.Delivery, func(), error) {
	// Check if the connection is healthy
	if r.connection == nil || r.connection.IsClosed() {
		return nil, nil, fmt.Errorf("connection is not open or is nil")
	}

	// Create a new channel
	ch, err := r.connection.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting channel: %w", err)
	}

	// Capture channel closure notifications
	closeErrChan := make(chan *amqp.Error)
	ch.NotifyClose(closeErrChan)

	// Log the channel creation with caller info
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown_file"
		line = 0
	}
	caller := fmt.Sprintf("%s:%d", file, line)
	logrus.WithField("caller", caller).Info("channel created")

	// Define a closer function to allow the caller to close the channel
	closer := func() {
		if ch != nil && !ch.IsClosed() {
			if err := ch.Close(); err != nil {
				logrus.WithField("name", name).WithField("caller", caller).WithError(err).Error("error closing channel")
			} else {
				logrus.WithField("name", name).WithField("caller", caller).Info("channel closed")
			}
		}
	}

	// Monitor channel closure in a goroutine for logging/debugging purposes
	go func() {
		if closeErr := <-closeErrChan; closeErr != nil {
			logrus.WithFields(logrus.Fields{
				"caller": caller,
				"name":   name,
			}).WithError(closeErr).Error("channel closed unexpectedly")
		}
	}()

	// Declare the queue
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		true,      // delete when unused
		exclusive, // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		closer()
		return nil, nil, fmt.Errorf("error declaring queue: %w", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,       // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		closer()
		return nil, nil, fmt.Errorf("error binding queue: %w", err)
	}

	// Start consuming messages
	deliveries, err := ch.Consume(
		q.Name,            // queue
		"",                // consumer
		true,              // auto-ack
		exclusiveConsumer, // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		closer()
		return nil, nil, fmt.Errorf("error consuming from queue: %w", err)
	}

	return deliveries, closer, nil
}

func (r *tRabbit) CreateExchange(exchangeName string) error {
	ch, err := r.connection.Channel()
	if err != nil {
		if err.Error() == channelExhaustedError {
			panic(err)
		}
		return fmt.Errorf("error getting channel: %w", err)
	}
	defer func() {
		err := ch.Close()
		if err != nil {
			logrus.WithError(err).Error("error closing channel")
		} else {
			logrus.WithField("channel", ch).Info("channel closed")
		}
	}()

	return ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		false,        // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
}

func (r *tRabbit) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	ch, err := r.connection.Channel()
	if err != nil {
		if err.Error() == channelExhaustedError {
			panic(err)
		}
		return fmt.Errorf("error getting channel: %w", err)
	}
	defer func() {
		err := ch.Close()
		if err != nil {
			logrus.WithError(err).Error("error closing channel")
		} else {
			logrus.WithField("channel", ch).Info("channel closed")
		}
	}()

	err = ch.Publish(exchange, key, mandatory, immediate, msg)
	if err != nil {
		return fmt.Errorf("error publishing message: %w", err)
	}

	return nil
}

func (r *tRabbit) CloseConnection() {
	logrus.Info("Closing rabbit connection...")
	err := r.connection.Close()
	if err != nil {
		logrus.Error(err)
	}
}
