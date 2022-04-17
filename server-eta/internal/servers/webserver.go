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

func isClosed(ch <-chan string) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
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
				if isClosed(messageChannel) {
					break
				}

				err := conn.WriteMessage(websocket.TextMessage, []byte(<-messageChannel))
				if err != nil {
					log.Println("write failed:", err)
				}
			}
		}()

		defer func() {
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
