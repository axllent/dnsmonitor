package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/xconstruct/go-pushbullet"
)

// SendNotifications will send a notification to all registered apps
func SendNotifications(title, message, priority string) {
	go NotifyPushbullet(title, message, priority)
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

// NotifyPushbullet messages pushbullet
func NotifyPushbullet(title, message, priority string) {
	if config.PushbulletToken != "" {
		pb := pushbullet.New(config.PushbulletToken)

		if config.PushbulletDevice == "" {
			devs, err := pb.Devices()
			if err != nil {
				fmt.Println("Pushbullet:", err)
				return
			}
			if len(devs) == 0 {
				fmt.Println("No devices appear to be set up for Pushbullet")
				os.Exit(2)
			}
			config.PushbulletDevice = devs[0].Nickname
			fmt.Printf("Pushbullet: pushbullet_device not set in config, using \"%s\"\n", config.PushbulletDevice)
		}

		dev, err := pb.Device(config.PushbulletDevice)
		if err != nil {
			fmt.Println("Pushbullet:", err)
			return
		}

		err = dev.PushNote(title, message)
		if err != nil {
			fmt.Println("Pushbullet:", err)
			return
		}
	}
}
