package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var (
	websocketUpgrader = websocket.Upgrader{}
)

type Coordinates struct {
	BusId int32
	X     float64
	Y     float64
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func createAMQPChannel() (*amqp.Channel, *amqp.Connection) {
	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

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

func setupWebsocket(path string, consumer func(int, string, *websocket.Conn), sendFunc *(func(int, string))) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		*sendFunc = func(msgType int, msg string) {
			err := conn.WriteMessage(msgType, []byte(msg))
			if err != nil {
				log.Println("write failed:", err)
			}
		}

		defer func() {
			*sendFunc = nil
		}()

		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if msgType != -1 {
					log.Println("read failed:", err)
				}
				break
			}
			if consumer != nil {
				consumer(msgType, string(message), conn)
			}
		}
	})
}

func encodeObject(v any) []byte {
	enc, err := json.Marshal(v)
	failOnError(err, "Failed to marshal message")
	return enc
}

func presentData(data *[]Coordinates) string {
	return string(encodeObject(*data))
}

func setupWebServer(sendFunc *(func(int, string))) {
	setupWebsocket("/timing", nil, sendFunc)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/public/websockets.html")
	})

	go func() {
		http.ListenAndServe(":80", nil)
	}()
}

func setupMessageQueueListener(queueName string, consumer func(string)) {
	ch, conn := createAMQPChannel()

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

func main() {
	flag.Parse()

	data := []Coordinates{}

	var sendFunc func(int, string)
	setupWebServer(&sendFunc)

	setupMessageQueueListener("vehicle-coordinates", func(msg string) {
		coords := Coordinates{}
		json.Unmarshal([]byte(msg), &coords)
		data = append(data, coords)

		if sendFunc != nil {
			sendFunc(websocket.TextMessage, presentData(&data))
		}
	})

	<-make(chan bool)
}
