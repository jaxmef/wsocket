package main

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jaxmef/wsocket"
)

func main() {
	resolver := wsocket.NewJSONResolver("type")
	resolver.AddHandler("sum-response", handleSum)

	wsClient := wsocket.NewClient(context.Background(), resolver, nil, 10)
	wsClient.AddMiddleware(messageLogger)

	serverURL := "ws://localhost:8080/ws"
	u, err := url.Parse(serverURL)
	if err != nil {
		log.Printf("Failed to parse URL: %v\n", err)
		return
	}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Failed to connect to server: %v\n", err)
		return
	}
	conn := wsClient.NewConnection(c)
	go func() {
		err = conn.WriteMessage(wsocket.NewTextMessage([]byte(`{"type": "sum-request", "a": 1, "b": 2}`)))
		if err != nil {
			log.Printf("Failed to write message: %v\n", err)
			return
		}
		time.Sleep(1 * time.Second)
		err = conn.Close()
		if err != nil {
			log.Printf("Failed to close connection: %v\n", err)
			return
		}
	}()
	<-conn.Wait()
	log.Printf("Connection closed")
}

type sumResult struct {
	Type   string `json:"type"`
	Result int    `json:"result"`
}

func handleSum(_ context.Context, msg []byte, _ wsocket.ResponseWriter) error {
	jsonMsg := &sumResult{}
	if err := json.Unmarshal(msg, jsonMsg); err != nil {
		return err
	}

	log.Printf("Sum result: %d\n", jsonMsg.Result)

	return nil
}

func messageLogger(ctx context.Context, msg []byte) (context.Context, []byte, error) {
	log.Printf("Received message: %s\n", string(msg))
	return ctx, msg, nil
}
