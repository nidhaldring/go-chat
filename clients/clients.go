package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"nhooyr.io/websocket"
)

var r = redis.NewClient(&redis.Options{})

type Client struct {
	id     string
	conn   *websocket.Conn
	chatId string
}

func NewClient(conn *websocket.Conn, chatId string) *Client {
	c := &Client{conn: conn, id: randomName(), chatId: chatId}

	return c
}

func (c *Client) Write(ctx context.Context, e event) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	msg, err := e.MarshalBinary()
	if err != nil {
		return err
	}

	return c.conn.Write(ctx, websocket.MessageText, msg)
}

// Send random created client name to frontend app
// @TODO: ofc this is not a good idea please change this
func (c *Client) SendClientName(ctx context.Context) error {
	e := event{SenderId: c.id, Type: "init"}
	return c.Write(ctx, e)
}

func (c *Client) listenToMsg(ctx context.Context) ([]byte, error) {
	// @TODO: properly handle timeout here, it does not make sense to cancel conn
	// if the user didn't send a message but is still active on chat
	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	_, msg, err := c.conn.Read(ctx)
	return msg, err
}

func (c *Client) PublishOwnMessages(
	ctx context.Context,
	errCh chan error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := c.listenToMsg(ctx)
			if err != nil {
				errCh <- err
				return
			}

			// @TODO: handle the case when connection is closed
			s := r.Publish(ctx, c.chatId, event{SenderId: c.id, Msg: string(data), Type: Normal})
			log.Println(s)
		}
	}
}

func (c *Client) ListenToOtherMessages(
	ctx context.Context,
	errCh chan error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sub := r.Subscribe(ctx, c.chatId)
	c.broadcastStatusUpdate(ctx, r)

	var event event
	ch := sub.Channel()
	for {
		select {
		case msg := <-ch:
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				errCh <- err
				return
			}

			if event.Type == Status || event.SenderId != c.id {
				c.Write(ctx, event)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) broadcastStatusUpdate(ctx context.Context, r *redis.Client) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	res := r.PubSubNumSub(ctx, c.chatId)
	count := res.Val()[c.chatId]

  // @TODO: add "active since" to the status
	msg := fmt.Sprintf("Members count %d", count)
	r.Publish(ctx, c.chatId, event{Msg: msg, Type: Status})
}

type eventType string

const (
	Init   eventType = "init"
	Normal eventType = "normal"
	Status eventType = "status"
)

type event struct {
	Type     eventType `json:"type"`
	Msg      string    `json:"msg"`
	SenderId string    `json:"sender"`
}

func (e event) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}
