package main

import (
    "log"
    "net/http"

    "socialnet/service"
)

func main() {
    srv := service.NewServer()
    log.Println("starting server on :8080")
    if err := http.ListenAndServe(":8080", srv); err != nil {
        log.Fatal(err)
    }
}
