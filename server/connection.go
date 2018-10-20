package server

import (
	"net"
	"sync"
	"time"
)

// Connection wraps net.Conn but counts the total number of written bytes
type Connection struct {
	net.Conn
	sync.Mutex

	Started   time.Time
	SessionID string
	Remote    string

	written int
}

func (c *Connection) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.Lock()
	c.written += n
	c.Unlock()
	return n, err
}

func (c *Connection) Written() int {
	c.Lock()
	value := c.written
	c.Unlock()
	return value
}
