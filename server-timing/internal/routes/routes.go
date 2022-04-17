package routes

type Route struct {
	Id                 int32
	Stations           []Station
	AvgTimeBtwStations []float64 // Time in minutes
}

type Station struct {
	Id int32
	X  float64
	Y  float64
}

var (
	routesByBus = make(map[int32]Route)
)

func GetRouteForBus(busId int32) (Route, bool) {
	if val, ok := routesByBus[busId]; ok {
		return val, true
	} else {
		return Route{}, false
	}
}
