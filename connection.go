package wsocket

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type Connection interface {
	// Write writes a message to the connection.
	// If the message type is 0, it is set to websocket.TextMessage.
	// If the message type is not websocket.TextMessage, websocket.BinaryMessage or websocket.CloseMessage, an error is returned.
	Write(message Message) error

	// Close closes the connection.
	// If the connection is already closed, an error is returned.
	Close() error

	// Wait returns a channel that is closed when the connection is closed.
	Wait() <-chan struct{}
}

type connection struct {
	logger Logger

	conn       *websocket.Conn
	closedChan chan struct{}
}

func (c *connection) Write(message Message) error {
	if message.msgType == 0 {
		message.msgType = websocket.TextMessage
	}
	switch message.msgType {
	case websocket.TextMessage, websocket.BinaryMessage, websocket.CloseMessage:
		// Do nothing
	default:
		return fmt.Errorf("invalid message type: %d", message.msgType)
	}

	err := c.conn.WriteMessage(message.msgType, message.Message)
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	return nil
}

func (c *connection) Close() error {
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("failed to close connection:", err)
	}

	return nil
}

func (c *connection) Wait() <-chan struct{} {
	return c.closedChan
}
