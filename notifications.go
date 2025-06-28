package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// SendNotifications will send a notification to all registered apps
func sendNotifications(title, message, priority string) {
	go notifyGotify(title, message, priority)
}

// NotifyGotify messages gotify
func notifyGotify(title, message, priority string) {
	if config.GotifyServer != "" && config.GotifyToken != "" {
		u, err := url.Parse(config.GotifyServer)
		if err != nil {
			fmt.Println("Error parsing Gotify URL:", err.Error())
			return
		}

		u.Path = path.Join(u.Path, "message")
		queryString := u.Query()
		queryString.Set("token", config.GotifyToken)
		u.RawQuery = queryString.Encode()
		s := u.String()

		_, err = http.PostForm(s,
			url.Values{"message": {message}, "title": {title}, "priority": {priority}})
		if err != nil {
			fmt.Println("Gotify error:", err.Error())
			return
		}
	}
}
