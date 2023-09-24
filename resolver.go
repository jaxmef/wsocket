package wsocket

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/valyala/fastjson"
)

// Resolver is used to resolve a message to a handler.
type Resolver interface {
	// Handle resolves a message to a handler and returns the result of the handler.
	// If the message cannot be resolved, an error is returned.
	Handle(ctx context.Context, msg []byte, rw ResponseWriter) error
}

type Handler func(ctx context.Context, msg []byte, rw ResponseWriter) error

type JSONResolver struct {
	mu sync.RWMutex

	field    string
	handlers map[string]Handler
}

// NewJSONResolver creates a new JSONResolver instance.
// field is the name of the field in the JSON message that is used to resolve the handler.
// For example, if field is "type", the message {"type": "sum-request", "a": 1, "b": 2} is resolved to the handler registered for "sum-request".
// If the field is nested, use dot notation, e.g. "type.name".
func NewJSONResolver(field string) *JSONResolver {
	return &JSONResolver{
		field:    field,
		handlers: make(map[string]Handler),
	}
}

// AddHandler adds a handler for a message type.
// name is the value of the field that is used to resolve the handler.
func (r *JSONResolver) AddHandler(name string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = handler
}

func (r *JSONResolver) Handle(ctx context.Context, msg []byte, rw ResponseWriter) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fieldValue := fastjson.GetString(msg, strings.Split(r.field, ".")...)
	if fieldValue == "" {
		return fmt.Errorf("failed to get field %q from message", r.field)
	}

	handler, ok := r.handlers[fieldValue]
	if !ok {
		return fmt.Errorf("unknown message type '%q'", fieldValue)
	}

	return handler(ctx, msg, rw)
}
