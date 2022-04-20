package algorithm

import (
	"errors"
	"fmt"
	"math"

	"server-eta/internal/routes"
	"server-eta/internal/vehicles"
)

type AlgorithmState struct {
	Routes            routes.RouteList
	Vehicles          vehicles.VehicleList
	StationsWithTimes map[int]StationWithTimes
}

type Coordinates struct {
	VehicleId int
	X         float64
	Y         float64
}

type StationWithTimes struct {
	Station routes.Station
	Times   map[int]float64
}

func getRouteForVehicle(vehicleId int, st AlgorithmState) (routes.Route, bool) {
	if routeId, ok := st.Vehicles.RoutesByVehicle[vehicleId]; ok {
		return st.Routes.GetRouteById(routeId)
	} else {
		return routes.Route{}, false
	}
}

func calcDistanceFromVehicleToRoute(station1 routes.Station, station2 routes.Station, vehicle vehicles.Vehicle) (float64, float64, float64, error) {
	stationToStationDeltaX := station2.X - station1.X
	stationToStationDeltaY := station2.Y - station1.Y

	stationToVehicleDeltaX := vehicle.X - station1.X
	stationToVehicleDeltaY := vehicle.Y - station1.Y

	dotProduct := stationToStationDeltaX*stationToVehicleDeltaX + stationToStationDeltaY*stationToVehicleDeltaY
	lengthSquared := math.Pow(stationToStationDeltaX, 2) + math.Pow(stationToStationDeltaY, 2)

	if lengthSquared == 0 {
		return 0, 0, 0, errors.New("the two stations are at the exact same location")
	}

	coefficient := dotProduct / lengthSquared

	var closestPointOnLineX float64
	var closestPointOnLineY float64

	if coefficient < 0 {
		closestPointOnLineX = station1.X
		closestPointOnLineY = station1.Y
	} else if coefficient > 1 {
		closestPointOnLineX = station2.X
		closestPointOnLineY = station2.Y
	} else {
		closestPointOnLineX = station1.X + coefficient*stationToStationDeltaX
		closestPointOnLineY = station1.Y + coefficient*stationToStationDeltaY
	}

	diffX := vehicle.X - closestPointOnLineX
	diffY := vehicle.Y - closestPointOnLineY

	return math.Sqrt(math.Pow(diffX, 2) + math.Pow(diffY, 2)), closestPointOnLineX, closestPointOnLineY, nil
}

func UpdateTiming(coords Coordinates, state AlgorithmState) ([]int, error) {
	vehicle, ok := state.Vehicles.GetById(coords.VehicleId)

	if !ok {
		return []int{}, errors.New(fmt.Sprintf("vehicle with id %d not found", coords.VehicleId))
	}

	vehicle.X = coords.X
	vehicle.Y = coords.Y
	//state.Vehicles[coords.VehicleId] = vehicle

	route, found := getRouteForVehicle(coords.VehicleId, state)

	if !found {
		return []int{}, errors.New(fmt.Sprintf("cannot found route for vehicle with id %d", coords.VehicleId))
	}

	getStationByIndex := func(i int) routes.Station {
		station, _ := state.Routes.GetStationById(route.Stations[i])
		return station
	}

	if len(route.Stations) < 2 {
		return []int{}, errors.New(fmt.Sprintf("route %d has less than 2 stations", route.Id))
	}

	var startingStationIx int

	if vehicle.LastStationKnown {
		for i := 0; i < len(route.Stations); i++ {
			if route.Stations[i] == vehicle.LastStation {
				startingStationIx = i
			}
		}
	} else {
		startingStationIx = 0
	}

	shortestDistance := 0.0
	shortestDistanceIx := -1
	var closestPointOnLineX float64
	var closestPointOnLineY float64

	for i := startingStationIx; i < len(route.Stations)-1; i++ {
		distanceFromVehicleToRoute, pointOnLineX, pointOnLineY, err := calcDistanceFromVehicleToRoute(getStationByIndex(i), getStationByIndex(i+1), vehicle)
		if err != nil {
			continue
		}

		if shortestDistanceIx == -1 || distanceFromVehicleToRoute < shortestDistance {
			shortestDistance = distanceFromVehicleToRoute
			shortestDistanceIx = i
			closestPointOnLineX = pointOnLineX
			closestPointOnLineY = pointOnLineY
		}
	}

	distanceToNextStopX := getStationByIndex(shortestDistanceIx+1).X - closestPointOnLineX
	distanceToNextStopY := getStationByIndex(shortestDistanceIx+1).Y - closestPointOnLineY
	distanceToNextStop := math.Sqrt(math.Pow(distanceToNextStopX, 2) + math.Pow(distanceToNextStopY, 2))
	lengthCurrentStretchX := getStationByIndex(shortestDistanceIx+1).X - getStationByIndex(shortestDistanceIx).X
	lengthCurrentStretchY := getStationByIndex(shortestDistanceIx+1).Y - getStationByIndex(shortestDistanceIx).Y
	lengthCurrentStretch := math.Sqrt(math.Pow(lengthCurrentStretchX, 2) + math.Pow(lengthCurrentStretchY, 2))
	fractionOfCurrentStrecthLeft := distanceToNextStop / lengthCurrentStretch
	timeLeftCurrentStretch := route.AvgTimeBetweenStations[shortestDistanceIx] * fractionOfCurrentStrecthLeft

	timeToNextStation := timeLeftCurrentStretch

	var updatedStations []int

	for i := shortestDistanceIx + 1; i < len(route.Stations); i++ {
		if station, ok := state.StationsWithTimes[route.Stations[i]]; ok {
			station.Times[vehicle.Id] = timeLeftCurrentStretch
			state.StationsWithTimes[route.Stations[i]] = station
		} else {
			obj := StationWithTimes{Station: getStationByIndex(i), Times: map[int]float64{vehicle.Id: timeLeftCurrentStretch}}
			state.StationsWithTimes[route.Stations[i]] = obj
		}
		updatedStations = append(updatedStations, route.Stations[i])
		if i < len(route.Stations)-1 {
			timeToNextStation += route.AvgTimeBetweenStations[i]
		}
	}

	return updatedStations, nil
}
