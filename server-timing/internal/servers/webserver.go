package servers

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{}
)

func setupWebsocket(path string, consumer func(int, string, *websocket.Conn), sendFunc *(func(int, string))) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		defer conn.Close()

		*sendFunc = func(msgType int, msg string) {
			err := conn.WriteMessage(msgType, []byte(msg))
			if err != nil {
				log.Println("write failed:", err)
			}
		}

		defer func() {
			*sendFunc = nil
		}()

		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if msgType != -1 {
					log.Println("read failed:", err)
				}
				break
			}
			if consumer != nil {
				consumer(msgType, string(message), conn)
			}
		}
	})
}

func SetupWebServer(sendFunc *(func(int, string))) {
	setupWebsocket("/timing", nil, sendFunc)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/public/websockets.html")
	})

	go func() {
		http.ListenAndServe(":80", nil)
	}()
}
