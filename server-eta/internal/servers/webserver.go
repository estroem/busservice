package servers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{}
)

type NewChannel struct {
	URI     string
	Channel chan string
}

func setupWebsocket(path string) <-chan NewChannel {
	newConnectionChannel := make(chan NewChannel)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		messageChannel := make(chan string)

		newConnectionChannel <- NewChannel{URI: r.RequestURI, Channel: messageChannel}

		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		go func() {
			for {
				log.Default().Println("ready to send")

				msg, open := <-messageChannel
				if !open {
					log.Default().Println("message channel is closed")
					break
				}

				log.Default().Printf("sending message %s", msg)

				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("write failed:", err)
				}
			}
		}()

		defer func() {
			log.Default().Println("closing message channel")
			close(messageChannel)
		}()

		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if msgType != -1 {
					log.Println("read failed:", err)
				}
				break
			}
			log.Default().Println("writing to channel")
			messageChannel <- string(message)
		}
	})

	return newConnectionChannel
}

func SetupWebServer() <-chan NewChannel {
	newConnChan := setupWebsocket("/eta")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/public/websockets.html")
	})

	go func() {
		http.ListenAndServe(":80", nil)
	}()

	return newConnChan
}
