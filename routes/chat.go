package routes

import (
	"go-chat/manager"
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

var clientManager = manager.NewManager()

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
	chatId := req.PathValue("id")
	conn, err := websocket.Accept(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.CloseNow()

	client := manager.NewClient(conn)
	clientManager.Append(client, chatId)

	client.Init()
	for {
		data, err := client.ListenToMsg()
		if err != nil {
			clientManager.Remove(client, chatId)
			log.Println(err)
			break
		}

		clientManager.WriteClientMsg(client, chatId, data)
	}
}
