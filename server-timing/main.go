package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"

	algorithm "server-timing/internal/algorithm"
	"server-timing/internal/routes"
	server "server-timing/internal/servers"
)

type WebsocketChannel struct {
	StationId int32
	Channel   chan string
}

func isClosed(ch <-chan string) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func convertChannel(channel server.NewChannel) (WebsocketChannel, error) {
	regex, _ := regexp.Compile("stationId=([0-9]+)")
	stationIdStr := regex.FindStringSubmatch(channel.URI)
	if len(stationIdStr) != 2 || stationIdStr[1] == "" {
		return WebsocketChannel{}, errors.New(fmt.Sprintf("could not find stationId in URI %s", channel.URI))
	}
	stationId, _ := strconv.Atoi(stationIdStr[1])
	return WebsocketChannel{StationId: int32(stationId), Channel: channel.Channel}, nil
}

func sliceContains(item int32, slice []int32) bool {
	for _, el := range slice {
		if el == item {
			return true
		}
	}
	return false
}

func presentData(times map[int32]float64) string {
	str := ""
	for busId, minutesUntilArrival := range times {
		str += fmt.Sprintf("BusId %d: %f minutes until arrival\n", busId, minutesUntilArrival)
	}
	return str
}

func initState() algorithm.AlgorithmState {
	station1 := routes.Station{
		Id: 1,
		X:  0,
		Y:  0,
	}
	station2 := routes.Station{
		Id: 3,
		X:  0,
		Y:  1,
	}
	route := routes.Route{
		Id:                 1,
		Stations:           []routes.Station{station1, station2},
		AvgTimeBtwStations: []float64{1},
	}
	buses := map[int32]algorithm.Bus{
		3: {
			Id:               3,
			X:                0,
			Y:                0,
			Route:            route,
			LastStation:      routes.Station{},
			LastStationKnown: false,
		},
	}
	return algorithm.AlgorithmState{
		Buses:             buses,
		StationsWithTimes: map[int32]algorithm.StationWithTimes{},
	}
}

func main() {
	flag.Parse()

	messageChannels := []WebsocketChannel{}
	state := initState()

	newConnectionChannel := server.SetupWebServer()

	go func() {
		for {
			newChannel, err := convertChannel(<-newConnectionChannel)
			if err != nil {
				log.Default().Print(err)
				continue
			}
			messageChannels = append(messageChannels, newChannel)
			log.Default().Printf("new connection with stationId: %d\n", newChannel.StationId)
		}
	}()

	rabbitmq_username := GetConfig("rabbitmq_username")
	rabbitmq_password := GetConfig("rabbitmq_password")

	server.SetupMessageQueueListener(rabbitmq_username, rabbitmq_password, "vehicle-coordinates", func(msg string) {
		coords := algorithm.Coordinates{}
		json.Unmarshal([]byte(msg), &coords)

		updatedStations, err := algorithm.UpdateTiming(coords, state)
		if err != nil {
			log.Println(err)
			return
		}

		for i := 0; i < len(messageChannels); i++ {
			if !isClosed(messageChannels[i].Channel) {
				if sliceContains(messageChannels[i].StationId, updatedStations) {
					if times, found := state.StationsWithTimes[messageChannels[i].StationId]; found {
						messageChannels[i].Channel <- presentData(times.Times)
					}
				}
			} else {
				messageChannels = append(messageChannels[:i], messageChannels[i+1:]...)
				i--
			}
		}
	})

	<-make(chan bool)
}
