package main

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)
		wwwHost, _ := appengine.ModuleHostname(ctx, "www", "", "")
		var redirectURL string
		if appengine.IsDevAppServer() {
			redirectURL = fmt.Sprintf("//%s/%s", wwwHost, r.URL.Path)
		} else {
			redirectURL = fmt.Sprintf("https://www.%s%s", r.Host, r.URL.Path)
		}
		log.Debugf(ctx, "Redirecting to: %s", redirectURL)
		http.Redirect(w, r, redirectURL, 302)
	})
	appengine.Main()
}
