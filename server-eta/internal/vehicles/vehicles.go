package vehicles

import (
	"encoding/json"
	"errors"
	"fmt"
	"server-eta/internal/routes"
	"server-eta/internal/servers"

	"github.com/streadway/amqp"
)

type VehicleListInterface interface {
	GetRouteForVehicle(int) (routes.Route, bool)
}

type VehicleList struct {
	Vehicles        []Vehicle
	RoutesByVehicle map[int]int
}

type Vehicle struct {
	Id               int
	X                float64
	Y                float64
	LocationKnown    bool
	LastStation      int
	LastStationKnown bool
}

var (
	runningVehiclesRequestQueue amqp.Queue
	runningVehiclesQueue        amqp.Queue
	queuesCreated               bool = false
)

func (vl VehicleList) GetById(id int) (Vehicle, bool) {
	for _, vehicle := range vl.Vehicles {
		if id == vehicle.Id {
			return vehicle, true
		}
	}
	return Vehicle{}, false
}

func (vl VehicleList) UpdateCoords(id int, x float64, y float64) error {
	vehicle, found := vl.GetById(id)
	if !found {
		return errors.New(fmt.Sprintf("vehicle with id %d not found", id))
	}
	vehicle.X = x
	vehicle.Y = y
	return nil
}

func FetchRunnngVehicles() VehicleList {
	if !queuesCreated {
		runningVehiclesRequestQueue = servers.CreateQueue("vehicles-running-request")
		runningVehiclesQueue = servers.CreateQueue("vehicles-running")
		queuesCreated = true
	}

	servers.SendMessage("", runningVehiclesRequestQueue)

	routesByVehicle := make(map[int]int)
	json.Unmarshal([]byte(<-servers.Listen(runningVehiclesQueue.Name)), &routesByVehicle)

	vehicles := make([]Vehicle, len(routesByVehicle))
	i := 0
	for key, _ := range routesByVehicle {
		vehicles[i] = Vehicle{
			Id:            key,
			X:             0.0,
			Y:             0.0,
			LocationKnown: false,
		}
		i++
	}

	return VehicleList{Vehicles: vehicles, RoutesByVehicle: routesByVehicle}
}
