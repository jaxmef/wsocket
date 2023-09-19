package wsocket

import "github.com/gorilla/websocket"

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
