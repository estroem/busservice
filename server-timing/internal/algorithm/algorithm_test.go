package algorithm

import (
	"math"
	"server-timing/internal/routes"
	"testing"
)

const DELTA = 0.001

func withinDelta(a float64, b float64) bool {
	return math.Abs(a-b) < DELTA
}

func TestCalcDistanceFromBusToRoute(t *testing.T) {
	station1 := routes.Station{
		Id: 1,
		X:  0,
		Y:  0,
	}
	station2 := routes.Station{
		Id: 2,
		X:  1,
		Y:  1,
	}
	route := routes.Route{
		Id:                 1,
		Stations:           []routes.Station{station1, station2},
		AvgTimeBtwStations: []float64{1},
	}
	bus := Bus{
		Id:               1,
		X:                0.5,
		Y:                0.5,
		Route:            route,
		LastStation:      routes.Station{},
		LastStationKnown: false,
	}

	a, b, c, err := calcDistanceFromBusToRoute(station1, station2, bus)

	t.Logf("a=%f, b=%f, c=%f\n", a, b, c)

	if err != nil {
		t.Error(err)
	}

	if !withinDelta(a, 0) {
		t.Errorf("a=%f, expected 0", a)
	}

	if !withinDelta(b, 0.5) {
		t.Errorf("b=%f, expected 0.5", b)
	}

	if !withinDelta(c, 0.5) {
		t.Errorf("c=%f, expected 0.5", c)
	}
}

func TestCalcDistanceFromBusToRoute2(t *testing.T) {
	station1 := routes.Station{
		Id: 1,
		X:  0,
		Y:  0,
	}
	station2 := routes.Station{
		Id: 2,
		X:  0,
		Y:  1,
	}
	route := routes.Route{
		Id:                 1,
		Stations:           []routes.Station{station1, station2},
		AvgTimeBtwStations: []float64{1},
	}
	bus := Bus{
		Id:               1,
		X:                0.5,
		Y:                0.5,
		Route:            route,
		LastStation:      routes.Station{},
		LastStationKnown: false,
	}

	a, b, c, err := calcDistanceFromBusToRoute(station1, station2, bus)

	t.Logf("a=%f, b=%f, c=%f\n", a, b, c)

	if err != nil {
		t.Error(err)
	}

	if !withinDelta(a, 0.5) {
		t.Errorf("a=%f, expected 0.5", a)
	}

	if !withinDelta(b, 0) {
		t.Errorf("b=%f, expected 0", b)
	}

	if !withinDelta(c, 0.5) {
		t.Errorf("c=%f, expected 0.5", c)
	}
}

func TestUpdateTiming(t *testing.T) {
	station1 := routes.Station{
		Id: 1,
		X:  0,
		Y:  0,
	}
	station2 := routes.Station{
		Id: 2,
		X:  0,
		Y:  1,
	}
	route := routes.Route{
		Id:                 1,
		Stations:           []routes.Station{station1, station2},
		AvgTimeBtwStations: []float64{1},
	}
	bus := Bus{
		Id:               1,
		X:                0.5,
		Y:                0.5,
		Route:            route,
		LastStation:      routes.Station{},
		LastStationKnown: false,
	}

	state := AlgorithmState{
		Buses:             map[int32]Bus{bus.Id: bus},
		StationsWithTimes: map[int32]StationWithTimes{},
	}

	coords := Coordinates{
		BusId: bus.Id,
		X:     bus.X,
		Y:     bus.Y,
	}

	updatedStations, err := UpdateTiming(coords, state)

	t.Logf("updatedStations=%+v\n", updatedStations)
	t.Logf("state=%+v\n", state)

	if err != nil {
		t.Error(err)
	}

	if len(updatedStations) != 1 {
		t.Errorf("len(updatedStations)=%d, expected 1", len(updatedStations))
	}

	if len(updatedStations) == 1 && updatedStations[0] != 2 {
		t.Errorf("updated stationd id=%d, expected 2", updatedStations[0])
	}

	if len(state.StationsWithTimes) != 1 {
		t.Errorf("len(state.StationsWithTimes)=%d, expected 1", len(state.StationsWithTimes))
	}

	if len(state.StationsWithTimes) == 1 && withinDelta(state.StationsWithTimes[0].Times[0], 0.5) {
		t.Errorf("time to next station = %f, expected 0.5", state.StationsWithTimes[0].Times[0])
	}
}
