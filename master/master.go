package master

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

const defaultWSURL = "/master"

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	Token string
)

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(conn.RemoteAddr())
	defer conn.Close()

	var name string
	remoteHost, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	var msg = make(map[string]string)
	for {
		var err error
		if err = conn.ReadJSON(&msg); err != nil {
			break
		}

		log.Println(msg)
		switch msg["action"] {
		case "LOGIN":
			name = "http://" + remoteHost + ":" + msg["port"]
			err = conn.WriteJSON(map[string]string{
				"self": name,
			})
		case "HB":

		}
		if err != nil {
			break
		}
	}
}

func InitMaster(address, token string) {
	Token = token
	http.HandleFunc(defaultWSURL, WSHandler)
	log.Printf("Listening on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}
