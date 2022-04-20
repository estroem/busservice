package main

import (
	"log"

	"server-route/internal/server"
)

type RoutesListMessage struct {
	Routes   []Route
	Stations []Station
}

type Route struct {
	Id                     int
	Stations               []int
	AvgTimeBetweenStations []float64
}

type Station struct {
	Id int
	X  float64
	Y  float64
}

func initStations() []Station {
	station0 := Station{
		Id: 0,
		X:  0.1,
		Y:  0.1,
	}

	station1 := Station{
		Id: 1,
		X:  0.4,
		Y:  0.2,
	}

	station2 := Station{
		Id: 2,
		X:  0.1,
		Y:  0.5,
	}

	station3 := Station{
		Id: 3,
		X:  0.5,
		Y:  0.1,
	}

	station4 := Station{
		Id: 4,
		X:  0.9,
		Y:  0.5,
	}

	station5 := Station{
		Id: 5,
		X:  0.7,
		Y:  0.9,
	}

	station6 := Station{
		Id: 6,
		X:  0.3,
		Y:  0.9,
	}

	station7 := Station{
		Id: 6,
		X:  0.1,
		Y:  0.8,
	}

	return []Station{station0, station1, station2, station3, station4, station5, station6, station7}
}

func initRoutes() []Route {
	route0 := Route{
		Id:                     0,
		Stations:               []int{0, 1, 3, 4, 5},
		AvgTimeBetweenStations: []float64{2, 3, 2, 2},
	}

	route1 := Route{
		Id:                     1,
		Stations:               []int{4, 1, 2, 7},
		AvgTimeBetweenStations: []float64{2, 3, 4},
	}

	route2 := Route{
		Id:                     2,
		Stations:               []int{5, 6, 1, 3},
		AvgTimeBetweenStations: []float64{3, 2, 2},
	}

	route3 := Route{
		Id:                     3,
		Stations:               []int{7, 6, 4},
		AvgTimeBetweenStations: []float64{4, 3},
	}

	return []Route{route0, route1, route2, route3}
}

func main() {
	server.CreateChannel(GetConfig("rabbitmq_username"), GetConfig("rabbitmq_password"))
	defer server.CloseConnection()

	stations := initStations()
	routes := initRoutes()

	requestQueue := server.CreateQueue("routes-list-request")
	listQueue := server.CreateQueue("routes-list")

	msg := server.EncodeObject(RoutesListMessage{Routes: routes, Stations: stations})
	listener := server.Listen(requestQueue.Name)

	log.Default().Println("started server-route")

	for {
		log.Default().Println("listening for requests")
		<-listener
		log.Default().Println("received request")
		server.SendMessage(string(msg), listQueue)
		log.Default().Println("responsed to request")
	}
}
