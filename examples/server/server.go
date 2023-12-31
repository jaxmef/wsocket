package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jaxmef/wsocket"
)

func main() {
	resolver := wsocket.NewJSONResolver("type")
	resolver.AddHandler("sum-request", handleSum)

	wsClient := wsocket.NewClient(context.Background(), resolver, nil, 10)
	wsClient.AddMiddleware(messageLogger)

	upgrader := &websocket.Upgrader{}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection: %v\n", err)
			return
		}

		conn := wsClient.NewConnection(c)
		<-conn.Wait()
		log.Printf("Connection closed after %f seconds", time.Since(start).Seconds())
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("Failed to start server: %v\n", err)
	}
}

type sumRequest struct {
	Type string `json:"type"`
	A    int    `json:"a"`
	B    int    `json:"b"`
}

func handleSum(_ context.Context, msg []byte, rw wsocket.ResponseWriter) error {
	jsonMsg := &sumRequest{}
	if err := json.Unmarshal(msg, jsonMsg); err != nil {
		return err
	}

	result := jsonMsg.A + jsonMsg.B
	response := []byte(fmt.Sprintf(`{"type": "sum-response", "result": %d}`, result))
	if err := rw.WriteMessage(wsocket.NewTextMessage(response)); err != nil {
		return err
	}

	return nil
}

func messageLogger(ctx context.Context, msg []byte) (context.Context, []byte, error) {
	log.Printf("Received message: %s\n", string(msg))
	return ctx, msg, nil
}
