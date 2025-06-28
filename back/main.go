package main

import (
	"net/http"

	"socialnet/proto/posts"
	"socialnet/service"
)

func main() {
	srv := posts.NewPostServiceServer(service.NewServer())
	http.ListenAndServe(":8080", srv)
}
