package main

import (
	"encoding/json"
	"flag"
	"log"

	server "server-timing/internal/servers"
)

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

func main() {
	flag.Parse()

	data := []Coordinates{}
	messageChannels := []server.NewChannel{}

	newConnectionChannel := server.SetupWebServer()

	go func() {
		for {
			newChannel := <-newConnectionChannel
			messageChannels = append(messageChannels, newChannel)
			log.Default().Printf("new connection: %s\n", newChannel.URI)
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
