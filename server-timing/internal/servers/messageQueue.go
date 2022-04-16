package servers

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func createAMQPChannel(username string, password string) (*amqp.Channel, *amqp.Connection) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@definition.default.svc.cluster.local:5672/", username, password))
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

func SetupMessageQueueListener(username string, password string, queueName string, consumer func(string)) {
	ch, conn := createAMQPChannel(username, password)

	q := createAMQPQueue(queueName, ch)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		defer ch.Close()
		defer conn.Close()

		forever := make(chan bool)

		go func() {
			for d := range msgs {
				log.Printf("Received a message: %s", d.Body)
				consumer(string(d.Body))
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

		<-forever
	}()
}
