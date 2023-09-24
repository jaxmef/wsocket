package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jaxmef/wsocket"
)

// This example shows how to use a complex resolvers to handle different types of messages.
// The server will handle messages with the type "sum-request" and "event".
// "event" requests will be handled by a separate resolver.
// Event resolver will handle messages with the data.type "info" and "error".

func main() {
	wsClient := wsocket.NewClient(context.Background(), getResolver(), nil)
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

func getResolver() wsocket.Resolver {
	eventResolver := getEventResolver()

	resolver := wsocket.NewJSONResolver("type")
	resolver.AddHandler("sum-request", handleSum)
	resolver.AddHandler("event", eventResolver.Handle)

	return resolver
}

func getEventResolver() wsocket.Resolver {
	eventResolver := wsocket.NewJSONResolver("data.type")
	eventResolver.AddHandler("info", handleInfoEvent)
	eventResolver.AddHandler("error", handleErrorEvent)
	return eventResolver
}

func messageLogger(ctx context.Context, msg []byte) (context.Context, []byte, error) {
	log.Printf("Received message: %s\n", string(msg))
	return ctx, msg, nil
}
