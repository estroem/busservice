package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"

	server "server-timing/internal/servers"
)

type WebsocketChannel struct {
	StationId int32
	Channel   chan string
}

type Coordinates struct {
	BusId int32
	X     float64
	Y     float64
}

func encodeObject(v any) []byte {
	enc, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}
	return enc
}

func presentData(data *[]Coordinates) string {
	return string(encodeObject(*data))
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

func main() {
	flag.Parse()

	data := []Coordinates{}
	messageChannels := []WebsocketChannel{}

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
		coords := Coordinates{}
		json.Unmarshal([]byte(msg), &coords)
		data = append(data, coords)

		str := presentData(&data)

		for i := 0; i < len(messageChannels); i++ {
			if !isClosed(messageChannels[i].Channel) {
				messageChannels[i].Channel <- str
			} else {
				messageChannels = append(messageChannels[:i], messageChannels[i+1:]...)
				i--
			}
		}
	})

	<-make(chan bool)
}
