package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"server-vehicle/internal/server"
)

func startDriving(busId int32, routeId int32) {
}

func encodeObject(v any) []byte {
	enc, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}
	return enc
}

func main() {
	flag.Parse()

	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

	server.CreateChannel(rabbitmq_username, rabbitmq_password)
	defer server.CloseConnection()

	q := server.CreateQueue("vehicles")

	for {
		server.SendMessage("", q)
		time.Sleep(20 * time.Second)
	}
}
