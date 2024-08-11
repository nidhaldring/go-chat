package main

import (
	"fmt"
	"go-chat/routes"
	"net/http"
)

func main() {
	port := "5000"

	server := http.NewServeMux()

	// set up routes
	routes.SetUpChatRouters(server)
	routes.SetUpHomePageRouters(server)

	fmt.Printf("Listening on port %s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), server)
}
