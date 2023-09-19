package wsocket

import "github.com/gorilla/websocket"

// Message is a message that can be sent to a connection.
// It is recommended to use the NewTextMessage, NewBinaryMessage and NewCloseMessage functions to create a Message.
type Message struct {
	msgType int
	Message []byte
}

func NewTextMessage(msg []byte) Message {
	return Message{
		msgType: websocket.TextMessage,
		Message: msg,
	}
}

func NewBinaryMessage(msg []byte) Message {
	return Message{
		msgType: websocket.BinaryMessage,
		Message: msg,
	}
}

func NewCloseMessage() Message {
	return Message{
		msgType: websocket.CloseMessage,
	}
}
