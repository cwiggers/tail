package slave

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/cwiggers/bitcask"
	"github.com/gorilla/websocket"
	log "github.com/laohanlinux/go-logger/logger"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	wsclient *websocket.Conn
	bcStore  *bitcask.BitCask
)

func storeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Bad Request", 400)
			return
		}
		value, err := bcStore.Get([]byte(key))
		if err != nil {
			if err == bitcask.ErrNotFound {
				http.Error(w, "Not Found", 404)
				return
			} else {
				log.Warn("failed to get value, Err:[%s]", err)
				http.Error(w, "Bad Gateway", 502)
				return
			}
		}
		fmt.Fprintf(w, string(value))
	case "PUT":
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Bad Request", 400)
			return
		}

		value, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Warn("failed to get body, Err:[%s]", err)
			http.Error(w, "Bad Request", 400)
		}
		err = bcStore.Put([]byte(key), value)
		if err != nil {
			log.Warn("failed to set value, Err:[%s]", err)
			http.Error(w, "Bad Gateway", 502)
			return
		}
	case "DELETE":
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Bad Request", 400)
			return
		}
		err := bcStore.Del([]byte(key))
		if err != nil {
			log.Warn("failed to del key, Err:[%s]", err)
			http.Error(w, "Bad Gateway", 502)
			return
		}
	default:
		http.Error(w, "Method not allowed", 405)
		return
	}
}

func InitSlave(bcDir, masterAddr, slaveAddr, token string) (err error) {
	// init bitcask
	bc, err := bitcask.Open(bcDir, nil)
	if err != nil {
		log.Fatal(err)
	}
	bcStore = bc

	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return
	}

	req, _ := json.Marshal(map[string]string{
		"action": "LOGIN",
		"token":  token,
	})
	if _, err = conn.Write(req); err != nil {
		return err
	}

	// Listen peers update
	go func() {
		for {
			req := json.Marshal(
				map[string]string{
					"action": "HB",
					"token":  token,
				})
			if _, err := conn.Write(req); err != nil {

			}

			if err != nil {
				log.Warn("Connection to master closed, retry in 10 seconds")
				time.Sleep(time.Second * 10)
				InitSlave(bcDir, masterAddr, slaveAddr, token)
				break
			}
			time.Sleep(time.Second)
		}
	}()

	http.HandleFunc("/store", storeHandler)
	http.ListenAndServe(slaveAddr, nil)

	bcStore.Close()
	return nil
}
