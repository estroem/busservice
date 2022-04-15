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

func setupWebsocket(path string, consumer func(int, string, *websocket.Conn)) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		for {
			log.Default().Println("waiting for message")
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read failed:", err)
				break
			}
			consumer(msgType, string(message), conn)
		}
		log.Default().Println("exiting websocket function")
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

func setupWebServer(data *[]Coordinates) {
	setupWebsocket("/timing", func(msgType int, msg string, conn *websocket.Conn) {
		err := conn.WriteMessage(msgType, []byte(presentData(data)))
		if err != nil {
			log.Println("write failed:", err)
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/public/websockets.html")
	})

	go func() {
		http.ListenAndServe(":80", nil)
	}()
}

func setupMessageQueueListener(data *[]Coordinates) {
	log.Printf("here3")
	ch, conn := createAMQPChannel()

	log.Printf("here3")
	q := createAMQPQueue("vehicle-coordinates", ch)

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

	log.Printf("here")
	go func() {
		log.Printf("here2")
		defer ch.Close()
		defer conn.Close()

		forever := make(chan bool)

		go func() {
			for d := range msgs {
				log.Printf("Received a message: %s", d.Body)
				data2 := Coordinates{}
				json.Unmarshal(d.Body, &data2)
				*data = append(*data, data2)
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

		<-forever
	}()
}

func main() {
	flag.Parse()

	data := []Coordinates{}

	setupWebServer(&data)
	setupMessageQueueListener(&data)

	<-make(chan bool)
}
