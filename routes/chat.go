package routes

import (
	"context"
	"encoding/json"
	"go-chat/clients"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/redis/go-redis/v9"
	"nhooyr.io/websocket"
)

func SetUpChatRouters(server *http.ServeMux) {
	server.Handle("GET /chat/{id}", http.HandlerFunc(renderChatPage))
	server.Handle("GET /chat/connect/{id}", http.HandlerFunc(handleChatStart))
}

var r = redis.NewClient(&redis.Options{})

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
	conn, err := websocket.Accept(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.CloseNow()

	chatId := req.PathValue("id")
	client := clients.NewClient(conn)
	client.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error)
	go publishOwnMessages(ctx, client, chatId, r, errCh)
	go listenToOtherMessages(ctx, client, chatId, r, errCh)

	error := <-errCh
	log.Printf("connection failed with err: %s\n", error)
}

type Message struct {
	ClientId, Message string
}

func (m Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func publishOwnMessages(
	ctx context.Context,
	client *clients.Client,
	chatId string,
	r *redis.Client,
	errCh chan error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			data, err := client.ListenToMsg(ctx)
			if err != nil {
				errCh <- err
				return
			}

			// @TODO: handle the case when connection is closed
			s := r.Publish(ctx, chatId, Message{ClientId: client.GetId(), Message: string(data)})
			log.Println(s)
		}
	}
}

func listenToOtherMessages(
	ctx context.Context,
	currentClient *clients.Client,
	chatId string,
	r *redis.Client,
	errCh chan error,
) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var payload Message
	sub := r.Subscribe(ctx, chatId)
	ch := sub.Channel()
	for {
		select {
		case msg := <-ch:
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err != nil {
				errCh <- err
				return
			}

			if payload.ClientId != currentClient.GetId() {
				currentClient.Write(ctx, payload.Message, payload.ClientId)
			}
		case <-ctx.Done():
			return
		}
	}
}
