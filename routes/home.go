package routes

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
)

func SetUpHomePageRouters(server *http.ServeMux) {
	server.Handle("GET /", http.HandlerFunc(handleStaticFiles))
}

func handleStaticFiles(res http.ResponseWriter, req *http.Request) {
	filePath := path.Join(".", "static", req.URL.Path)

	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.NotFound(res, req)
			return
		}
		log.Fatal(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	if info.IsDir() {
		filePath = path.Join(filePath, "index.html")
		_, err := os.Stat(filePath)
		if err != nil {
			http.NotFound(res, req)
			return
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	res.Write(data)
}
