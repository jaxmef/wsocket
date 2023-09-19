package wsocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	resolver := NewJSONResolver("type")
	logger := &NoLogger{}

	c := NewClient(context.Background(), resolver, logger)

	assert.Equal(t, resolver, c.(*client).resolver)
	assert.Equal(t, logger, c.(*client).logger)
}

func TestClient_AddMiddleware(t *testing.T) {
	resolver := NewJSONResolver("type")

	c := NewClient(context.Background(), resolver, &NoLogger{})

	middleware := func(ctx context.Context, msg []byte) (context.Context, []byte, error) {
		return ctx, msg, nil
	}

	c.AddMiddleware(middleware)

	assert.Equal(t, 1, len(c.(*client).middlewares))
	valueOfExpectedMiddleware := reflect.ValueOf(middleware)
	valueOfActualMiddleware := reflect.ValueOf(c.(*client).middlewares[0])
	assert.True(t, valueOfExpectedMiddleware.Pointer() == valueOfActualMiddleware.Pointer())
}

func TestClient_NewConnection_Empty(t *testing.T) {
	resolver := NewJSONResolver("type")

	c := NewClient(context.Background(), resolver, &NoLogger{})

	conn := c.NewConnection(nil)
	assert.Nil(t, conn)
}

func TestHandleConnection_NewConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		defer conn.Close()

		message := []byte("Hello, client!")
		err = conn.WriteMessage(websocket.TextMessage, message)
		assert.NoError(t, err)

		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		assert.NoError(t, err)
	}))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer clientConn.Close()

	resolver := &testResolver{
		expectedMessage: []byte("Hello, client!"),
		t:               t,
	}

	client := NewClient(context.Background(), resolver, &NoLogger{})
	conn := client.NewConnection(clientConn)

	<-conn.Wait()

	assert.Equal(t, 1, resolver.calls)
}

func TestHandleConnection_NewConnection_WithMiddleware(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		defer conn.Close()

		message := []byte("Hello, client!")
		err = conn.WriteMessage(websocket.TextMessage, message)
		assert.NoError(t, err)

		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		assert.NoError(t, err)
	}))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer clientConn.Close()

	resolver := &testResolver{
		expectedMessage: []byte("Hello, client!"),
		t:               t,
	}

	client := NewClient(context.Background(), resolver, &NoLogger{})
	middlewareCalls := 0
	client.AddMiddleware(func(ctx context.Context, msg []byte) (context.Context, []byte, error) {
		middlewareCalls++
		return ctx, msg, nil
	})
	conn := client.NewConnection(clientConn)

	<-conn.Wait()

	assert.Equal(t, 1, resolver.calls)
	assert.Equal(t, 1, middlewareCalls)
}

func TestHandleConnection_NewConnection_WriteResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		assert.NoError(t, err)
		defer conn.Close()

		message := []byte(`{"type": "sum-request", "a": 1, "b": 2}`)
		err = conn.WriteMessage(websocket.TextMessage, message)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		msgType, msg, err := conn.ReadMessage()
		assert.Equal(t, websocket.TextMessage, msgType)
		assert.Equal(t, []byte(`{"type": "sum-response", "result": 3}`), msg)

		err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		assert.NoError(t, err)
	}))
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1)
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer clientConn.Close()

	resolver := NewJSONResolver("type")
	resolverCalls := 0
	resolver.AddHandler("sum-request", func(_ context.Context, msg []byte) (Message, error) {
		resolverCalls++

		jsonMsg := &struct {
			A int `json:"a"`
			B int `json:"b"`
		}{}
		err := json.Unmarshal(msg, jsonMsg)
		assert.NoError(t, err)

		result := jsonMsg.A + jsonMsg.B
		response := []byte(fmt.Sprintf(`{"type": "sum-response", "result": %d}`, result))

		return NewTextMessage(response), nil
	})

	client := NewClient(context.Background(), resolver, &NoLogger{})
	conn := client.NewConnection(clientConn)

	<-conn.Wait()

	assert.Equal(t, 1, resolverCalls)
}

type testResolver struct {
	expectedMessage []byte
	t               *testing.T
	calls           int
}

func (r *testResolver) Handle(_ context.Context, msg []byte) (Message, error) {
	r.calls++
	assert.Equal(r.t, r.expectedMessage, msg)
	return Message{}, nil
}
