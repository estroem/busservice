package servers

import (
	"log"

	"github.com/gorilla/websocket"
)

func SetupWebClient(url string) chan []byte {
	log.Default().Printf("dialing url %s", url)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	ch := make(chan []byte)

	go func() {
		defer conn.Close()
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				if mt == -1 {
					log.Default().Println("closing connection")
					close(ch)
					return
				} else {
					log.Default().Print("error on read from websocket", err)
				}
			}
			ch <- message
		}
	}()

	return ch
}
