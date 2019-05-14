package redis

import (
	"encoding/json"
	"log"

	"github.com/chuhlomin/busnj-console/pkg/websocket"

	"github.com/mediocregopher/radix/v3"
)

const (
	busVehicleDataChannel = "busVehicleDataChannel"
)

// Client represents layer between writer and Redis
type Client struct {
	pool *radix.Pool
	conn radix.Conn
	ps   radix.PubSubConn
}

// NewClient creates new Client
func NewClient(network string, addr string, size int) (*Client, error) {
	pool, err := radix.NewPool(network, addr, size)
	if err != nil {
		return nil, err
	}

	conn, err := radix.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	ps := radix.PubSub(conn)

	return &Client{
		pool: pool,
		conn: conn,
		ps:   ps,
	}, nil
}

// LoadBusVehicleDataMessages all BusVehicleDataMessages
func (c *Client) LoadBusVehicleDataMessages() ([]*BusVehicleDataMessage, error) {
	var keys []string
	err := c.pool.Do(
		radix.Cmd(&keys, "KEYS", "busVehicleData:*"),
	)
	if err != nil {
		return nil, err
	}

	result := []*BusVehicleDataMessage{}
	for _, key := range keys {
		busVehicleData, err := c.loadBusVehicleData(key)
		if err != nil {
			log.Printf("Failed to load BusVehicleData by key: %s, err: %v", key, err)
			continue
		}

		result = append(result, busVehicleData)
	}

	return result, nil
}

// ConsumeBusVehicleDataChannel subscribes to Redis PubSub channel
// and forwards it to WebSocket clients
func (c *Client) ConsumeBusVehicleDataChannel(hub *websocket.Hub) {
	ch := make(chan radix.PubSubMessage)

	if err := c.ps.Subscribe(ch, busVehicleDataChannel); err != nil {
		panic(err)
	}

	for {
		select {
		case msg := <-ch:
			hub.Broadcast <- msg.Message
		}
	}
}

func (c *Client) loadBusVehicleData(key string) (*BusVehicleDataMessage, error) {
	var message string
	err := c.pool.Do(
		radix.Cmd(&message, "GET", key),
	)
	if err != nil {
		return nil, err
	}

	var busVehicleDataMessage BusVehicleDataMessage
	err = json.Unmarshal([]byte(message), &busVehicleDataMessage)
	if err != nil {
		return nil, err
	}

	return &busVehicleDataMessage, nil
}
