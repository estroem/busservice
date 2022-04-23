package main

import (
	"client-etasign/internal/servers"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("stationId not found in command line")
	}
	stationId, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("stationId not found in command line")
	}
	ch := servers.SetupWebClient(fmt.Sprintf("ws://busservice.info/web/eta?stationId=%d", stationId))
	fmt.Print("\033[H\033[2J")

	for msg := range ch {
		fmt.Print("\033[H\033[2J")
		mp := make(map[int]float64)
		json.Unmarshal(msg, &mp)
		for key, val := range mp {
			fmt.Printf("Bus %d:\t%.0f min\n", key, math.Round(val))
		}
	}
}
