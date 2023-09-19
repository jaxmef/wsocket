package wsocket

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

type Client interface {
	AddMiddleware(middleware Middleware)
	NewConnection(conn *websocket.Conn) Connection
}

type client struct {
	mu          sync.RWMutex
	middlewares []Middleware

	ctx      context.Context
	resolver Resolver
	logger   Logger
}

type Middleware func(ctx context.Context, msg []byte) (context.Context, []byte, error)

// NewClient creates a new client instance.
// ctx is used to cancel the client.
// resolver is used to resolve incoming messages.
// logger is used to log errors. If nil, a default logger is used. You can use NoLogger to disable logging.
func NewClient(ctx context.Context, resolver Resolver, logger Logger) Client {
	if logger == nil {
		logger = &defaultLogger{}
	}

	return &client{
		mu:          sync.RWMutex{},
		ctx:         ctx,
		resolver:    resolver,
		logger:      logger,
		middlewares: make([]Middleware, 0),
	}
}

// AddMiddleware adds a middleware to the client.
// Middlewares are executed in the order they are added.
func (s *client) AddMiddleware(middleware Middleware) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middlewares = append(s.middlewares, middleware)
}

// NewConnection creates a new connection instance.
// conn is used to read and write messages.
// If conn is nil, nil is returned.
// The connection is automatically closed when the client is canceled.
func (s *client) NewConnection(conn *websocket.Conn) Connection {
	if conn == nil {
		return nil
	}

	c := &connection{
		logger:     s.logger,
		conn:       conn,
		closedChan: make(chan struct{}),
	}

	go s.handleConnection(c)

	return c
}

func (s *client) handleConnection(c *connection) {
	defer func() {
		if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.logger.Printf("failed to close connection: %v", err)
		}
		close(c.closedChan)
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// todo: check type
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				if errors.Is(err, net.ErrClosed) || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				s.logger.Printf("failed to read message: %v", err)
				return
			}

			ctx, msg, err := s.runMiddlewares(msg)
			if err != nil {
				s.logger.Printf("failed to run middlewares: %v", err)
				continue
			}

			resp, err := s.resolver.Handle(ctx, msg)
			if err != nil {
				s.logger.Printf("failed to handle message: %v", err)
				continue
			}

			if resp.msgType != 0 {
				err = c.Write(resp)
				if err != nil {
					s.logger.Printf("failed to write message: %v", err)
					continue
				}
			}
		}
	}
}

func (s *client) runMiddlewares(msg []byte) (context.Context, []byte, error) {
	ctx := context.Background()

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, middleware := range s.middlewares {
		var err error
		ctx, msg, err = middleware(ctx, msg)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, msg, nil
}
