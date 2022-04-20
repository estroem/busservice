package routes

import (
	"encoding/json"
	"log"
	"server-eta/internal/servers"
)

type RouteListInterface interface {
	GetRouteById(int) (Route, bool)
}

type RouteList struct {
	routes   []Route
	stations []Station
}

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

var (
	routes   []Route
	stations []Station
)

func (rl RouteList) GetRouteById(routeId int) (Route, bool) {
	for _, route := range rl.routes {
		if routeId == route.Id {
			return route, true
		}
	}
	return Route{}, false
}

func (rl RouteList) GetStationById(id int) (Station, bool) {
	for _, station := range rl.stations {
		if id == station.Id {
			return station, true
		}
	}
	return Station{}, false
}

func FetchRoutes() RouteList {
	listReq := servers.CreateQueue("routes-list-request")
	list := servers.CreateQueue("routes-list")

	log.Default().Println("fetching routes: created queues")

	servers.SendMessage("", listReq)

	log.Default().Println("fetching routes: sent message")

	var routeListObj RoutesListMessage
	msg := <-servers.Listen(list.Name)
	json.Unmarshal([]byte(msg), &routeListObj)

	log.Default().Println("fetching routes: received response")

	routes = routeListObj.Routes
	stations = routeListObj.Stations

	return RouteList{routes: routes, stations: stations}
}
