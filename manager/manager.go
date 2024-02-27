package manager

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"nhooyr.io/websocket"
)

type ClientManager struct {
	mu      sync.Mutex
	clients map[string][]*Client
}

func NewManager() *ClientManager {
	return &ClientManager{
		mu:      sync.Mutex{},
		clients: make(map[string][]*Client),
	}
}

func (cm *ClientManager) Append(c *Client, id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.clients[id]; !ok {
		cm.clients[id] = append(make([]*Client, 0), c)
	} else {
		cm.clients[id] = append(cm.clients[id], c)
	}
}

func (cm *ClientManager) Remove(clientToRemove *Client, id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.clients[id] != nil {
		cm.clients[id] = filter(cm.clients[id], func(c *Client) bool { return clientToRemove != c })
	}
}

func (cm *ClientManager) WriteClientMsg(currentClient *Client, chatId string, msg []byte) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.clients[chatId] != nil {
		for _, c := range cm.clients[chatId] {
			if c != currentClient {
				err := c.Write(msg, currentClient)
				if err != nil {
					log.Print(err)
					cm.Remove(c, chatId)
				}
			}
		}
	}
}

type Client struct {
	id   string
	conn *websocket.Conn
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn, id: randomName()}
}

func (c *Client) Write(data []byte, sender *Client) error {
	msg, err := json.Marshal(struct {
		Msg    string `json:"msg"`
		Sender string `json:"sender"`
	}{Msg: string(data), Sender: sender.id})
	if err != nil {
		return err
	}

	return c.conn.Write(context.Background(), websocket.MessageText, msg)
}
