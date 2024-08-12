package routes

import (
	"context"
	"go-chat/clients"
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
	conn, err := websocket.Accept(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.CloseNow()

	chatId := req.PathValue("id")
	client := clients.NewClient(conn, chatId)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.SendClientName(ctx)

	errCh := make(chan error)
	go client.PublishOwnMessages(ctx, errCh)
	go client.ListenToOtherMessages(ctx, errCh)

	error := <-errCh
	log.Printf("connection closed/failed with err: %s\n", error)
}
