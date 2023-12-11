package wsocket

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type ResponseWriter interface {
	// WriteMessage writes a message to the connection.
	// If the message type is 0, it is set to websocket.TextMessage.
	// If the message type is not websocket.TextMessage, websocket.BinaryMessage or websocket.CloseMessage, an error is returned.
	WriteMessage(msg Message) error
}

type Connection interface {
	ResponseWriter

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

	writeChan chan Message
}

func newConnection(logger Logger, conn *websocket.Conn, writeChanSize int) *connection {
	c := &connection{
		logger:     logger,
		conn:       conn,
		closedChan: make(chan struct{}),
		writeChan:  make(chan Message, writeChanSize),
	}

	go c.messageWriter()

	return c
}

func (c *connection) WriteMessage(message Message) error {
	if message.msgType == 0 {
		message.msgType = websocket.TextMessage
	}
	switch message.msgType {
	case websocket.TextMessage, websocket.BinaryMessage, websocket.CloseMessage:
		// Do nothing
	default:
		return fmt.Errorf("invalid message type: %d", message.msgType)
	}

	c.writeChan <- message
	return nil
}

func (c *connection) messageWriter() {
	for {
		select {
		case <-c.closedChan:
			return
		case msg := <-c.writeChan:
			err := c.conn.WriteMessage(msg.msgType, msg.Message)
			if err != nil {
				c.logger.Printf("failed to write message: %v", err)
				return
			}
		}
	}
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
