package main

import (
	"net/http"

	"socialnet/proto/posts"
	"socialnet/service"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	postsSrv := service.NewPostsServer()
	authSrv := service.NewAuthServer()

	http.Handle(posts.PostServicePathPrefix, corsHandler(posts.NewPostServiceServer(postsSrv)))
	http.Handle(posts.AuthServicePathPrefix, corsHandler(posts.NewAuthServiceServer(authSrv)))

	http.ListenAndServe(":8080", nil)
}
