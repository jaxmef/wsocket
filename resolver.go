package wsocket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type Resolver interface {
	Handle(ctx context.Context, msg []byte) (Message, error)
}

type Handler func(ctx context.Context, msg []byte) (Message, error)

type JSONResolver struct {
	mu sync.Mutex

	field    string
	handlers map[string]Handler
}

func NewJSONResolver(field string) *JSONResolver {
	return &JSONResolver{
		field:    field,
		handlers: make(map[string]Handler),
	}
}

func (r *JSONResolver) AddHandler(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = handler
}

func (r *JSONResolver) Handle(ctx context.Context, msg []byte) (Message, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg, &data); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	fieldValue, ok := getField(data, r.field)
	if !ok {
		return Message{}, fmt.Errorf("failed to get field %q from message", r.field)
	}

	handler, ok := r.handlers[fieldValue]
	if !ok {
		return Message{}, fmt.Errorf("unknown message type '%q'", fieldValue)
	}

	return handler(ctx, msg)
}

func getField(data map[string]interface{}, fieldPath string) (string, bool) {
	fields := strings.Split(fieldPath, ".")
	curr := data

	for i, field := range fields {
		value, ok := curr[field]
		if !ok {
			return "", false
		}

		if i == len(fields)-1 {
			if strValue, ok := value.(string); ok {
				return strValue, true
			}
			return "", false
		}

		if nestedData, ok := value.(map[string]interface{}); ok {
			curr = nestedData
		} else {
			return "", false
		}
	}

	return "", false
}
