package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var client *mongo.Client

func main() {
	// Connect to MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/meme10"
	}

	client, _ := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))

	// HTTP routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from meme10/web! MongoDB connected successfully.")
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		client.Database("meme10").Collection("memes").InsertOne(context.Background(), map[string]interface{}{
			"name": "test",
		})
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		cursor, err := client.Database("meme10").Collection("memes").Find(context.Background(), bson.D{})
		if err != nil {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var result map[string]interface{}
			cursor.Decode(&result)
			fmt.Fprintf(w, "%v\n", result)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := client.Ping(ctx, nil)
		if err != nil {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "OK")
	})

	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
