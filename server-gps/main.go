package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"time"

	server "server-gps/internal/server"
)

type Coordinates struct {
	VehicleId int32
	X         float64
	Y         float64
}

func createRandomGPSCoords(vehicleId int32) Coordinates {
	x := rand.Float64()
	y := rand.Float64()
	return Coordinates{VehicleId: vehicleId, X: x, Y: y}
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

	ch, conn := server.CreateAMQPChannel(rabbitmq_username, rabbitmq_password)
	defer ch.Close()
	defer conn.Close()

	q := server.CreateAMQPQueue("vehicle-coordinates", ch)

	for {
		var busId int32 = rand.Int31n(20)
		server.SendMessage(encodeObject(createRandomGPSCoords(busId)), q, ch)
		time.Sleep(1 * time.Second)
	}
}
