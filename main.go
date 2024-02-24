package main

import (
	"go-chat/routes"
	"net/http"
)

func main() {

	server := http.NewServeMux()

	// set up routes
	routes.SetUpChatRouters(server)
	routes.SetUpHomePageRouters(server)

	http.ListenAndServe(":5000", server)
}
