package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			// Format the function name to show only the last part
			function = frame.Function[strings.LastIndex(frame.Function, "/")+1:]
			// Format the file path to show only the file name and line number
			file = fmt.Sprintf("%s:%d", frame.File[strings.LastIndex(frame.File, "/")+1:], frame.Line)
			return function, file
		},
	})
}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Load the RabbitMQ instance
	rabbit := GetInstance()
	defer rabbit.CloseConnection()

	queueName := "exclusive-test-queue"
	routingKey := "test-key"
	exchangeName := "test-exchange"
	exclusive := true
	exclusiveConsumer := true

	// First connection to the exclusive queue
	deliveries, closer1, err := rabbit.ConsumeFromQueue(queueName, routingKey, exchangeName, exclusive, exclusiveConsumer, "ch 01")
	if err != nil {
		logrus.WithError(err).Error("First connection failed")
		return
	}
	logrus.Info("First consumer connected successfully to the exclusive queue")
	defer func() {
		closer1()
	}()

	// Attempt a second connection to the same exclusive queue (expected to fail)
	_, closer2, err := rabbit.ConsumeFromQueue(queueName, routingKey, exchangeName, exclusive, exclusiveConsumer, "ch 02")
	if err != nil {
		logrus.WithError(err).Error("Second connection failed as expected")
		return
	} else {
		logrus.Error("Second consumer connected successfully (unexpected behavior)")
	}
	defer func() {
		closer2()
	}()

	// Process messages from the first consumer for testing
	go func() {
		for delivery := range deliveries {
			logrus.WithField("name", "ch 01").Infof("Message received: %s", string(delivery.Body))
		}
	}()

	logrus.Info("Test completed")
}
