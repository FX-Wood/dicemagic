package main

import (
	"fmt"
	"net/http"

	"log"

	"cloud.google.com/go/logging"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/slack/roll/", SlackRollHandler)
	http.HandleFunc("/dflow/", DialogueWebhookHandler)
	http.HandleFunc("/chart/", drawChart)

	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "dice-magic-dev"

	// Creates a client.
	client, err := logging.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Sets the name of the log to write to.
	logName := "dicemagic-api"

	Debuglogger := client.Logger(logName).StandardLogger(logging.Debug)

	// Logs "hello world", log entry is visible at
	// Stackdriver Logs.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wwwHost, _ := appengine.ModuleHostname(ctx, "www", "", "")
		var redirectURL string
		if appengine.IsDevAppServer() {
			redirectURL = fmt.Sprintf("//%s/%s", wwwHost, r.URL.Path)
		} else {
			redirectURL = fmt.Sprintf("https://www.%s%s", r.Host, r.URL.Path)
		}
		Debuglogger.Printf("Redirecting to: %s", redirectURL)
		http.Redirect(w, r, redirectURL, 302)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
