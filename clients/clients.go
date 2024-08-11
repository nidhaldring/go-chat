package clients

import (
	"context"
	"encoding/json"
	"time"

	"nhooyr.io/websocket"
)

type event struct {
	Type   string `json:"type"`
	Msg    string `json:"msg"`
	Sender string `json:"sender"`
}

// @TODO: replace this feature with redis
// func (cm *ClientManager) Append(c *Client, chatId string) {
// 	cm.mu.Lock()
// 	defer cm.mu.Unlock()
// 	if _, ok := cm.clients[chatId]; !ok {
// 		cm.clients[chatId] = append(make([]*Client, 0), c)
// 	} else {
// 		cm.clients[chatId] = append(cm.clients[chatId], c)
// 	}

// 	cm.notifyAllSession(
// 		fmt.Sprintf("members: %d,active since: %s", len(cm.clients[chatId]), cm.activeSince.Format(time.ANSIC)),
// 		chatId,
// 	)
// }

type Client struct {
	id   string
	conn *websocket.Conn
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn, id: randomName()}
}

func (c *Client) Write(ctx context.Context, data string, senderId string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	msg, err := json.Marshal(event{Msg: data, Sender: senderId, Type: "message"})
	if err != nil {
		return err
	}

	return c.conn.Write(ctx, websocket.MessageText, msg)
}

func (c *Client) ListenToMsg(ctx context.Context) ([]byte, error) {
	// @TODO: properly handle timeout here, it does not make sense to cancel conn
	// if the user didn't send a message but is still active on chat
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	_, msg, err := c.conn.Read(ctx)
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

func (c *Client) GetId() string {
	return c.id
}
