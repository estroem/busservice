package util

import (
	"encoding/json"
	"log"
)

func IsClosed(ch chan string) bool {
	select {
	case x, ok := <-ch:
		if ok {
			ch <- x
			return false
		} else {
			return true
		}
	default:
		return false
	}
}

func EncodeObject(v any) []byte {
	enc, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal message: %s", err)
	}
	return enc
}

func SliceContains(item int, slice []int) bool {
	for _, el := range slice {
		if el == item {
			return true
		}
	}
	return false
}
