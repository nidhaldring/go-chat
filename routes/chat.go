package routes

import (
	"context"
	"encoding/json"
	"go-chat/utils"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"nhooyr.io/websocket"
)

func SetUpChatRouters(server *http.ServeMux) {
	server.Handle("GET /chat/{id}", http.HandlerFunc(renderChatPage))
	server.Handle("GET /chat/connect/{id}", http.HandlerFunc(handleChatStart))
}

func renderChatPage(res http.ResponseWriter, req *http.Request) {
	data, err := os.ReadFile(path.Join("templates", "chat.templ.html"))
	if err != nil {
		log.Println(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
	}

	chatTempl, err := template.New("chat").Parse(string(data))
	if err != nil {
		log.Println(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
	}

	chatTempl.Execute(res, struct{ Id string }{req.PathValue("id")})
}

func handleChatStart(res http.ResponseWriter, req *http.Request) {
	// @TODO: verify chatId is a valid uuid to prevent overflow issue etc...
	chatId := req.PathValue("id")
	conn, err := websocket.Accept(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// @TODO: investigate if we should use just "Close"
	defer conn.CloseNow()

	currentClient := NewClient(conn)
	_, ok := clients[chatId]
	if !ok {
		clients[chatId] = append(make([]*Client, 0), currentClient)
	} else {
		clients[chatId] = append(clients[chatId], currentClient)
	}

	// also returns if loop should end or not
	handleErr := func(err error) bool {
		if err == nil {
			return false
		}

		log.Println(err)
		clients[chatId] = utils.Filter(clients[chatId], func(elm *Client) bool { return elm.conn != conn })
		return true
	}
	for {
		_, data, err := conn.Read(context.Background())
		if handleErr(err) {
			break
		}

		chatMembers := clients[chatId]
		for _, m := range chatMembers {
			if m != currentClient {
				err := m.Write(data, currentClient)
				if handleErr(err) {
					break
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
	return &Client{conn: conn, id: utils.RandomName()}
}

func (c *Client) Write(data []byte, sender *Client) error {
	msg, err := json.Marshal(struct {
		Msg    string `json:"msg"`
		Sender string `json:"sender"`
  }{Msg:string(data), Sender: sender.id})
	if err != nil {
		return err
	}

	return c.conn.Write(context.Background(), websocket.MessageText, msg)
}

// @TODO: make this thread safe
var clients = make(map[string][]*Client)
