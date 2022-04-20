package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"server-eta/internal/algorithm"
	"server-eta/internal/routes"
	"server-eta/internal/servers"
	"server-eta/internal/util"
	"server-eta/internal/vehicles"
)

type WebsocketChannel struct {
	StationId int
	Channel   chan string
}

func presentData(data *map[int]float64) string {
	return string(util.EncodeObject(*data))
}

func convertChannel(channel servers.NewChannel) (WebsocketChannel, error) {
	regex, _ := regexp.Compile("stationId=([0-9]+)")
	stationIdStr := regex.FindStringSubmatch(channel.URI)
	if len(stationIdStr) != 2 || stationIdStr[1] == "" {
		return WebsocketChannel{StationId: -1, Channel: channel.Channel}, errors.New(fmt.Sprintf("could not find stationId in URI %s", channel.URI))
	}
	stationId, _ := strconv.Atoi(stationIdStr[1])
	return WebsocketChannel{StationId: stationId, Channel: channel.Channel}, nil
}

func main() {
	servers.CreateChannel(GetConfig("rabbitmq_username"), GetConfig("rabbitmq_password"))
	defer servers.CloseConnection()

	log.Default().Println("creates channel to rabbitmq")

	routeList := routes.FetchRoutes()

	log.Default().Println("fetched routes")

	vehicleList := vehicles.FetchRunnngVehicles()

	log.Default().Println("fetched vehicles")

	state := algorithm.AlgorithmState{
		Routes:            routeList,
		Vehicles:          vehicleList,
		StationsWithTimes: map[int]algorithm.StationWithTimes{},
	}

	messageChannels := []WebsocketChannel{}

	newConnectionChannel := servers.SetupWebServer()

	log.Default().Println("listening for websockets")

	go func() {
		for {
			newChannel, err := convertChannel(<-newConnectionChannel)
			if err != nil {
				close(newChannel.Channel)
				log.Default().Print(err)
				continue
			}
			messageChannels = append(messageChannels, newChannel)
			log.Default().Printf("new connection with stationId: %d\n", newChannel.StationId)
			if times, found := state.StationsWithTimes[newChannel.StationId]; found {
				log.Default().Printf("found station in state")
				newChannel.Channel <- presentData(&times.Times)
			}
		}
	}()

	coordQueue := servers.CreateQueue("vehicle-coordinates")

	for msg := range servers.Listen(coordQueue.Name) {
		coords := algorithm.Coordinates{}
		json.Unmarshal([]byte(msg), &coords)
		log.Default().Printf("received coordinates %+v", coords)

		updatedStations, err := algorithm.UpdateTiming(coords, state)
		if err != nil {
			log.Println(err)
			continue
		}

		for i := 0; i < len(messageChannels); i++ {
			if !util.IsClosed(messageChannels[i].Channel) {
				log.Default().Printf("channel with index %d is open", i)
				if util.SliceContains(messageChannels[i].StationId, updatedStations) {
					log.Default().Printf("channel with index %d is updated", i)
					if times, found := state.StationsWithTimes[messageChannels[i].StationId]; found {
						log.Default().Printf("channel with index %d has data", i)
						messageChannels[i].Channel <- presentData(&times.Times)
					}
				}
			} else {
				messageChannels = append(messageChannels[:i], messageChannels[i+1:]...)
				i--
			}
		}
	}

	<-make(chan bool)
}
