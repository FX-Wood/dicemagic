package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/logging"
)

func main() {
	http.HandleFunc("/dice-api", DiceAPIHandler)

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "k8s-dice-magic"
	redirectURL := "https://www.smallnet.org/"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Sets the name of the log to write to.
	logName := "dicemagic-api"

	Debuglogger := client.Logger(logName).StandardLogger(logging.Debug)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Debuglogger.Printf("Redirecting to: %s", redirectURL)
		http.Redirect(w, r, redirectURL, 302)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
func DiceAPIHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}
