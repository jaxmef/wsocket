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
func (c *client) AddMiddleware(middleware Middleware) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.middlewares = append(c.middlewares, middleware)
}

// NewConnection creates a new connection instance.
// websocketConn is used to read and write messages.
// If websocketConn is nil, nil is returned.
// The connection is automatically closed when the client is canceled.
func (c *client) NewConnection(websocketConn *websocket.Conn) Connection {
	if websocketConn == nil {
		return nil
	}

	conn := &connection{
		logger:     c.logger,
		conn:       websocketConn,
		closedChan: make(chan struct{}),
	}

	go c.handleConnection(conn)

	return conn
}

func (c *client) handleConnection(conn *connection) {
	defer func() {
		if err := conn.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			c.logger.Printf("failed to close connection: %v", err)
		}
		close(conn.closedChan)
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, msg, err := conn.conn.ReadMessage()
			if err != nil {
				if errors.Is(err, net.ErrClosed) || websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					return
				}
				c.logger.Printf("failed to read message: %v", err)
				return
			}

			go c.handleMessage(msg, conn)
		}
	}
}

func (c *client) handleMessage(msg []byte, conn *connection) {
	ctx, msg, err := c.runMiddlewares(msg)
	if err != nil {
		c.logger.Printf("failed to run middlewares: %v", err)
		return
	}

	resp, err := c.resolver.Handle(ctx, msg)
	if err != nil {
		c.logger.Printf("failed to handle message: %v", err)
		return
	}

	if resp.msgType != 0 {
		err = conn.Write(resp)
		if err != nil {
			c.logger.Printf("failed to write message: %v", err)
			return
		}
	}
}

func (c *client) runMiddlewares(msg []byte) (context.Context, []byte, error) {
	ctx := context.Background()

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, middleware := range c.middlewares {
		var err error
		ctx, msg, err = middleware(ctx, msg)
		if err != nil {
			return ctx, nil, err
		}
	}

	return ctx, msg, nil
}
