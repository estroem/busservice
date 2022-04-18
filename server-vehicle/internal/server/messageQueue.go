package server

import (
	"fmt"
	"log"

	amqp "github.com/streadway/amqp"
)

var (
	connection *amqp.Connection
	channel    *amqp.Channel
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func CreateChannel(username string, password string) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@definition.default.svc.cluster.local:5672/", username, password))
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	connection = conn
	channel = ch
}

func CloseConnection() {
	if channel != nil {
		channel.Close()
	}
	if connection != nil {
		connection.Close()
	}
}

func CreateQueue(name string) amqp.Queue {
	args := make(amqp.Table)
	args["x-message-ttl"] = 60000

	q, err := channel.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	failOnError(err, "Failed to declare a queue")
	return q
}

func SendMessage(body string, q amqp.Queue) {
	err := channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
}

func Listen(queueName string) <-chan string {
	msgs, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to register a consumer")

	channel := make(chan string)
	go func() {
		for d := range msgs {
			channel <- string(d.Body)
		}
	}()
	return channel
}
