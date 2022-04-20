package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"sync"

	"server-vehicle/internal/server"
)

var (
	mapMutex sync.Mutex
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
	mapMutex.Lock()
	for {
		num := rand.Int31n(limit)
		if _, found := list[num]; !found {
			mapMutex.Unlock()
			return num
		}
	}
}

func getRandomIn(limit int32, list map[int32]int32) int32 {
	mapMutex.Lock()
	for {
		num := rand.Int31n(limit)
		if _, found := list[num]; found {
			mapMutex.Unlock()
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

	//startingQueue := server.CreateQueue("vehicles-starting")
	//stoppingQueue := server.CreateQueue("vehicles-stopping")

	var runningVehicles map[int32]int32 = make(map[int32]int32) //Key: vehicleId, value: routeId

	runningVehicles[0] = 2
	runningVehicles[1] = 0
	runningVehicles[2] = 3
	runningVehicles[3] = 2
	runningVehicles[4] = 1
	runningVehicles[5] = 0
	runningVehicles[6] = 0
	runningVehicles[7] = 3
	runningVehicles[8] = 1
	runningVehicles[9] = 1
	runningVehicles[10] = 0
	runningVehicles[11] = 3
	runningVehicles[12] = 2
	runningVehicles[13] = 1
	runningVehicles[14] = 3
	runningVehicles[15] = 0
	runningVehicles[16] = 2
	runningVehicles[17] = 1
	runningVehicles[18] = 3
	runningVehicles[19] = 0

	/*
		go func() {
			go func() {
				for {
					message := VehicleStarting{
						VehicleId: getRandomNotIn(30, runningVehicles),
						RouteId:   rand.Int31n(5),
					}

					mapMutex.Lock()
					runningVehicles[message.VehicleId] = message.RouteId
					mapMutex.Unlock()

					server.SendMessage(encodeObject(message), startingQueue)
					time.Sleep(20 * time.Second)
				}
			}()

			time.Sleep(70 * time.Second)

			for {
				message := VehicleStopping{
					VehicleId: getRandomIn(10, runningVehicles),
				}

				mapMutex.Lock()
				delete(runningVehicles, message.VehicleId)
				mapMutex.Unlock()

				server.SendMessage(encodeObject(message), stoppingQueue)
				time.Sleep(20 * time.Second)
			}
		}()
	*/

	broadcastQueue := server.CreateQueue("vehicles-running-request")
	runningQueue := server.CreateQueue("vehicles-running")

	msg := encodeObject(runningVehicles)
	listener := server.Listen(broadcastQueue.Name)

	log.Default().Println("started server-route")

	for {
		log.Default().Println("listening for requests")
		<-listener
		log.Default().Println("received request")
		server.SendMessage(string(msg), runningQueue)
		log.Default().Println("responsed to request")
	}
}
