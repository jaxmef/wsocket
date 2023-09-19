package wsocket

import (
	"context"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestJSONResolver_Handle(t *testing.T) {
	// Create a JSONResolver instance for testing
	resolver := NewJSONResolver("type")

	// Add some test handlers
	resolver.AddHandler("type1", func(ctx context.Context, msg []byte) (Message, error) {
		return NewTextMessage([]byte("Handler for type1")), nil
	})
	resolver.AddHandler("type2", func(ctx context.Context, msg []byte) (Message, error) {
		return NewBinaryMessage([]byte("Handler for type2")), nil
	})
	resolver.AddHandler("type3", func(ctx context.Context, msg []byte) (Message, error) {
		return NewCloseMessage(), nil
	})

	tests := []struct {
		name           string
		inputMessage   string
		expectedResult string
		expectedType   int
		expectedError  bool
	}{
		{
			name:           "Valid Message Type 1",
			inputMessage:   `{"type": "type1"}`,
			expectedResult: "Handler for type1",
			expectedType:   websocket.TextMessage,
			expectedError:  false,
		},
		{
			name:           "Valid Message Type 2",
			inputMessage:   `{"type": "type2"}`,
			expectedResult: "Handler for type2",
			expectedType:   websocket.BinaryMessage,
			expectedError:  false,
		},
		{
			name:           "Valid Message Type 3",
			inputMessage:   `{"type": "type3"}`,
			expectedResult: "",
			expectedType:   websocket.CloseMessage,
			expectedError:  false,
		},
		{
			name:           "Unknown Message Type",
			inputMessage:   `{"type": "unknown"}`,
			expectedResult: "",
			expectedError:  true,
		},
		{
			name:           "Invalid JSON",
			inputMessage:   `invalid_json`,
			expectedResult: "",
			expectedError:  true,
		},
		{
			name:           "Missing Field",
			inputMessage:   `{}`,
			expectedResult: "",
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message, err := resolver.Handle(context.Background(), []byte(test.inputMessage))
			if test.expectedError {
				assert.Error(t, err, "Expected an error")
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, test.expectedResult, string(message.Message), "Unexpected message type")
				assert.Equal(t, test.expectedType, message.msgType, "Unexpected message type")
			}
		})
	}
}

func TestGetField(t *testing.T) {
	data := map[string]interface{}{
		"type": "test",
		"nested": map[string]interface{}{
			"field1": "value1",
		},
	}

	tests := []struct {
		name           string
		fieldPath      string
		expectedValue  string
		expectedResult bool
	}{
		{
			name:           "Valid Field",
			fieldPath:      "type",
			expectedValue:  "test",
			expectedResult: true,
		},
		{
			name:           "Nested Field",
			fieldPath:      "nested.field1",
			expectedValue:  "value1",
			expectedResult: true,
		},
		{
			name:           "Invalid Field",
			fieldPath:      "nonexistent",
			expectedValue:  "",
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value, ok := getField(data, test.fieldPath)
			if test.expectedResult {
				assert.True(t, ok)
				assert.Equal(t, test.expectedValue, value)
			} else {
				assert.False(t, ok)
			}
		})
	}
}
