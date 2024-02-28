package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type event struct {
	Type   string `json:"type"`
	Msg    string `json:"msg"`
	Sender string `json:"sender"`
}

type ClientManager struct {
	mu          sync.Mutex
	clients     map[string][]*Client
	activeSince time.Time
}

func NewManager() *ClientManager {
	return &ClientManager{
		mu:          sync.Mutex{},
		clients:     make(map[string][]*Client),
		activeSince: time.Now(),
	}
}

func (cm *ClientManager) Append(c *Client, chatId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.clients[chatId]; !ok {
		cm.clients[chatId] = append(make([]*Client, 0), c)
	} else {
		cm.clients[chatId] = append(cm.clients[chatId], c)
	}

	cm.notifyAllSession(
		fmt.Sprintf("members: %d,active since: %s", len(cm.clients[chatId]), cm.activeSince.Format(time.ANSIC)),
		chatId,
	)
}

func (cm *ClientManager) Remove(clientToRemove *Client, id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.clients[id] != nil {
		cm.clients[id] = filter(cm.clients[id], func(c *Client) bool { return clientToRemove != c })
	}
}

func (cm *ClientManager) WriteClientMsg(owner *Client, chatId string, msg []byte) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.clients[chatId] != nil {
		for _, c := range cm.clients[chatId] {
			if c != owner {
				err := c.Write(msg, owner)
				cm.handleErr(err, c, chatId)
			}
		}
	}
}

func (cm *ClientManager) notifyAllSession(message, chatId string) error {
	msg, err := json.Marshal(event{Msg: message, Type: "status"})
	if err != nil {
		return err
	}

	for _, client := range cm.clients[chatId] {
		err := client.conn.Write(context.Background(), websocket.MessageText, msg)
		cm.handleErr(err, client, chatId)
	}

	return nil
}

func (cm *ClientManager) handleErr(err error, c *Client, chatId string) {
	if err != nil {
		log.Print(err)
		cm.Remove(c, chatId)
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
	msg, err := json.Marshal(event{Msg: string(data), Sender: sender.id, Type: "message"})
	if err != nil {
		return err
	}

	return c.conn.Write(context.Background(), websocket.MessageText, msg)
}

func (c *Client) ListenToMsg() ([]byte, error) {
	_, msg, err := c.conn.Read(context.Background())
	return msg, err
}

func (c *Client) Init() error {
	msg, err := json.Marshal(event{Sender: c.id, Type: "init"})
	if err != nil {
		return err
	}

	c.conn.Write(context.Background(), websocket.MessageText, msg)
	return nil
}
