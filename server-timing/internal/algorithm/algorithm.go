package algorithm

import (
	"errors"
	"fmt"
	"math"

	routes "server-timing/internal/routes"
)

type AlgorithmState struct {
	Buses             map[int32]Bus
	StationsWithTimes map[int32]StationWithTimes
}

type Bus struct {
	Id               int32
	X                float64
	Y                float64
	Route            routes.Route
	LastStation      routes.Station
	LastStationKnown bool
}

type Coordinates struct {
	BusId int32
	X     float64
	Y     float64
}

type StationWithTimes struct {
	Station routes.Station
	Times   map[int32]float64
}

func calcDistanceFromBusToRoute(station1 routes.Station, station2 routes.Station, bus Bus) (float64, float64, float64, error) {
	stationToStationDeltaX := station2.X - station1.X
	stationToStationDeltaY := station2.Y - station1.Y

	stationToBusDeltaX := bus.X - station1.X
	stationToBusDeltaY := bus.Y - station1.Y

	dotProduct := stationToStationDeltaX*stationToBusDeltaX + stationToStationDeltaY*stationToBusDeltaY
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

	diffX := bus.X - closestPointOnLineX
	diffY := bus.Y - closestPointOnLineY

	return math.Sqrt(math.Pow(diffX, 2) + math.Pow(diffY, 2)), closestPointOnLineX, closestPointOnLineY, nil
}

func UpdateTiming(coords Coordinates, state AlgorithmState) ([]int32, error) {
	bus, ok := state.Buses[coords.BusId]

	if !ok {
		return []int32{}, errors.New(fmt.Sprintf("bus with id %d not found", coords.BusId))
	}

	bus.X = coords.X
	bus.Y = coords.Y
	state.Buses[coords.BusId] = bus

	route := bus.Route

	if len(route.Stations) < 2 {
		return []int32{}, errors.New(fmt.Sprintf("route %d has less than 2 stations", route.Id))
	}

	var startingStationIx int

	if bus.LastStationKnown {
		for i := 0; i < len(route.Stations); i++ {
			if route.Stations[i].Id == bus.LastStation.Id {
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
		distanceFromBusToRoute, pointOnLineX, pointOnLineY, err := calcDistanceFromBusToRoute(route.Stations[i], route.Stations[i+1], bus)
		if err != nil {
			continue
		}

		if shortestDistanceIx == -1 || distanceFromBusToRoute < shortestDistance {
			shortestDistance = distanceFromBusToRoute
			shortestDistanceIx = i
			closestPointOnLineX = pointOnLineX
			closestPointOnLineY = pointOnLineY
		}
	}

	distanceToNextStopX := route.Stations[shortestDistanceIx+1].X - closestPointOnLineX
	distanceToNextStopY := route.Stations[shortestDistanceIx+1].Y - closestPointOnLineY
	distanceToNextStop := math.Sqrt(math.Pow(distanceToNextStopX, 2) + math.Pow(distanceToNextStopY, 2))
	lengthCurrentStretchX := route.Stations[shortestDistanceIx+1].X - route.Stations[shortestDistanceIx].X
	lengthCurrentStretchY := route.Stations[shortestDistanceIx+1].Y - route.Stations[shortestDistanceIx].Y
	lengthCurrentStretch := math.Sqrt(math.Pow(lengthCurrentStretchX, 2) + math.Pow(lengthCurrentStretchY, 2))
	fractionOfCurrentStrecthLeft := distanceToNextStop / lengthCurrentStretch
	timeLeftCurrentStretch := route.AvgTimeBtwStations[shortestDistanceIx] * fractionOfCurrentStrecthLeft

	timeToNextStation := timeLeftCurrentStretch

	var updatedStations []int32

	for i := shortestDistanceIx + 1; i < len(route.Stations); i++ {
		if station, ok := state.StationsWithTimes[route.Stations[i].Id]; ok {
			station.Times[bus.Id] = timeLeftCurrentStretch
			state.StationsWithTimes[route.Stations[i].Id] = station
		} else {
			obj := StationWithTimes{Station: route.Stations[i], Times: map[int32]float64{bus.Id: timeLeftCurrentStretch}}
			state.StationsWithTimes[route.Stations[i].Id] = obj
		}
		updatedStations = append(updatedStations, route.Stations[i].Id)
		if i < len(route.Stations)-1 {
			timeToNextStation += route.AvgTimeBtwStations[i]
		}
	}

	return updatedStations, nil
}
