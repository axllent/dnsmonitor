package main

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// SendNotifications will send a notification to all registered apps
func SendNotifications(title, message, priority string) {
	go NotifyGotify(title, message, priority)
}

// NotifyGotify messages gotify
func NotifyGotify(title, message, priority string) {
	if config.GofifyServer != "" && config.GofifyToken != "" {
		u, err := url.Parse(config.GofifyServer)
		if err != nil {
			fmt.Println("Gotify:", err)
			return
		}

		u.Path = path.Join(u.Path, "message")
		queryString := u.Query()
		queryString.Set("token", config.GofifyToken)
		u.RawQuery = queryString.Encode()
		s := u.String()

		_, err = http.PostForm(s,
			url.Values{"message": {message}, "title": {title}, "priority": {priority}})
		if err != nil {
			fmt.Println("Gotify:", err)
			return
		}
	}
}
