package wsocket

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestConnection_Write(t *testing.T) {
	message := Message{
		Message: []byte("Hello, World!"),
	}

	serverReceivedMessage := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		defer conn.Close()

		// Read the message from the WebSocket client
		time.Sleep(100 * time.Millisecond)
		_, msg, err := conn.ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, string(message.Message), string(msg))
		serverReceivedMessage = true

		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Convert the server URL to a WebSocket URL
	serverURL := "ws" + server.URL[4:]

	// Connect to the mock WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	wsConn := &connection{
		conn: conn,
	}

	err = wsConn.Write(message)
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, serverReceivedMessage)
}

func TestConnection_Close(t *testing.T) {
	serverClosed := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		defer conn.Close()

		// Read the message from the WebSocket client
		time.Sleep(100 * time.Millisecond)
		_, _, err = conn.ReadMessage()
		assert.True(t, websocket.IsCloseError(err, websocket.CloseNormalClosure))

		serverClosed = true
	}))
	defer server.Close()

	// Convert the server URL to a WebSocket URL
	serverURL := "ws" + server.URL[4:]

	// Connect to the mock WebSocket server
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	wsConn := &connection{
		conn: conn,
	}

	err = wsConn.Close()
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, serverClosed)
}
