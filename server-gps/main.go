package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func createAMQPChannel() (*amqp.Channel, *amqp.Connection) {
	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

	//setUpgRPCServer()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@definition.default.svc.cluster.local:5672/", rabbitmq_username, rabbitmq_password))
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	return ch, conn
}

func createAMQPQueue(name string, ch *amqp.Channel) amqp.Queue {
	args := make(amqp.Table)
	args["x-message-ttl"] = 60000

	q, err := ch.QueueDeclare(
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

func sendMessage(body []byte, q amqp.Queue, ch *amqp.Channel) {
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	failOnError(err, "Failed to publish a message")
}

type Coordinates struct {
	BusId int32
	X     float64
	Y     float64
}

func createRandomGPSCoords(busId int32) Coordinates {
	x := rand.Float64()
	y := rand.Float64()
	return Coordinates{BusId: busId, X: x, Y: y}
}

func encodeObject(v any) []byte {
	enc, err := json.Marshal(v)
	failOnError(err, "Failed to marshal message")
	return enc
}

func main() {
	flag.Parse()

	ch, conn := createAMQPChannel()
	defer ch.Close()
	defer conn.Close()

	q := createAMQPQueue("vehicle-coordinates", ch)

	for {
		var busId int32 = rand.Int31n(100)
		sendMessage(encodeObject(createRandomGPSCoords(busId)), q, ch)
		time.Sleep(20 * time.Second)
	}
}
