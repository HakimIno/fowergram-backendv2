package messaging

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// NATSClient implements messaging using NATS
type NATSClient struct {
	conn *nats.Conn
}

// NewNATSClient creates a new NATS client
func NewNATSClient(natsURL string) (*NATSClient, error) {
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSClient{
		conn: conn,
	}, nil
}

// Close closes the NATS connection
func (n *NATSClient) Close() {
	n.conn.Close()
}

// Publish publishes a message to a subject
func (n *NATSClient) Publish(subject string, data []byte) error {
	return n.conn.Publish(subject, data)
}

// Subscribe subscribes to a subject
func (n *NATSClient) Subscribe(subject string, handler func(msg []byte)) error {
	_, err := n.conn.Subscribe(subject, func(m *nats.Msg) {
		handler(m.Data)
	})
	return err
}
