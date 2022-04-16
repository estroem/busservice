package main

import (
	"encoding/json"
	"flag"
	"log"

	server "server-timing/internal/servers"

	"github.com/gorilla/websocket"
)

type Coordinates struct {
	BusId int32
	X     float64
	Y     float64
}

func encodeObject(v any) []byte {
	enc, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}
	return enc
}

func presentData(data *[]Coordinates) string {
	return string(encodeObject(*data))
}

func main() {
	flag.Parse()

	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

	data := []Coordinates{}

	var sendFunc func(int, string)
	server.SetupWebServer(&sendFunc)

	server.SetupMessageQueueListener(rabbitmq_username, rabbitmq_password, "vehicle-coordinates", func(msg string) {
		coords := Coordinates{}
		json.Unmarshal([]byte(msg), &coords)
		data = append(data, coords)

		if sendFunc != nil {
			sendFunc(websocket.TextMessage, presentData(&data))
		}
	})

	<-make(chan bool)
}
