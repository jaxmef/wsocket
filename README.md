WSocket: Golang WebSocket client/server library.
=========================================================

[![GoDoc](https://godoc.org/github.com/jaxmef/wsocket?status.svg)](http://godoc.org/github.com/jaxmef/wsocket)
[![Coverage Status](https://coveralls.io/repos/github/jaxmef/wsocket/badge.svg?branch=feature/coveralls)](https://coveralls.io/github/jaxmef/wsocket?branch=feature/coveralls)

The aim of this library is to provide a similar way to regular HTTP routing for WebSocket messages. This is done by using a message resolver. The default resolver is a JSON resolver that will match on a field in the JSON message. Custom resolvers can be created by implementing the `wsocket.Resolver` interface.

## Features
- JSON message resolver
- Custom message resolvers
- Text and binary message support
- Middleware support
- Context support

## Usage
More examples of usage can be found in the [examples](examples) directory.
```go
// Create a new resolver that will match on the "message.type" field
// JSONResolver supports nested fields. For example "message.type" will match on {"message": {"type": "value"}}
resolver := wsocket.NewJSONResolver("message.type")
// Add a handler for the "sum-request" type
resolver.AddHandler("sum-request", handleSum)

...

// Create a new client with the resolver and disable logger
wsClient := wsocket.NewClient(context.Background(), resolver, &wsocket.NoLogger{})
// Add a middleware that will log all messages
wsClient.AddMiddleware(messageLogger)

...

http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
	
    // upgrade the connection to a websocket connection using gorilla/websocket
    c, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Failed to upgrade connection: %v\n", err)
        return
    }
    
    // create a new connection using the client and the websocket connection
    conn := wsClient.NewConnection(c)
    <-conn.Wait()
    log.Printf("Connection closed after %f seconds", time.Since(start).Seconds())
})

...

// wsocket.Connection has a Write method that will write the message to the websocket connection
err = conn.WriteMessage(wsocket.NewTextMessage([]byte(`{"type": "sum-request", "a": 1, "b": 2}`)))
if err != nil {
    log.Printf("Failed to write message: %v\n", err)
    return
}
```
