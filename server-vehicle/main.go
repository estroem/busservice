package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"time"

	"server-vehicle/internal/server"
)

type VehicleStarting struct {
	VehicleId int32
	RouteId   int32
}

type VehicleStopping struct {
	VehicleId int32
}

func encodeObject(v any) string {
	enc, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}
	return string(enc)
}

func getRandomNotIn(limit int32, list map[int32]int32) int32 {
	for {
		num := rand.Int31n(limit)
		if _, found := list[num]; !found {
			return num
		}
	}
}

func getRandomIn(limit int32, list map[int32]int32) int32 {
	for {
		num := rand.Int31n(limit)
		if _, found := list[num]; found {
			return num
		}
	}
}

func main() {
	flag.Parse()

	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

	server.CreateChannel(rabbitmq_username, rabbitmq_password)
	defer server.CloseConnection()

	startingQueue := server.CreateQueue("vehicles-starting")
	stoppingQueue := server.CreateQueue("vehicles-stopping")

	var runningVehicles map[int32]int32 = make(map[int32]int32) //Key: vehicleId, value: routeId

	go func() {
		go func() {
			for {
				message := VehicleStarting{
					VehicleId: getRandomNotIn(30, runningVehicles),
					RouteId:   rand.Int31n(5),
				}
				runningVehicles[message.VehicleId] = message.RouteId
				server.SendMessage(encodeObject(message), startingQueue)
				time.Sleep(20 * time.Second)
			}
		}()

		time.Sleep(70 * time.Second)

		for {
			message := VehicleStopping{
				VehicleId: getRandomIn(10, runningVehicles),
			}
			delete(runningVehicles, message.VehicleId)
			server.SendMessage(encodeObject(message), stoppingQueue)
			time.Sleep(20 * time.Second)
		}
	}()

	broadcastQueue := server.CreateQueue("vehicles-running-request")
	runningQueue := server.CreateQueue("vehicles-running")

	for {
		<-server.Listen(broadcastQueue.Name)
		server.SendMessage(encodeObject(runningVehicles), runningQueue)
	}
}
