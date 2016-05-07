package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-xmpp"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var baseURL = "https://www.livecoding.tv/"
var client *xmpp.Client
var jidSuffix = "@livecoding.tv"
var roomSuffix = "@chat.livecoding.tv"
var usernameToFind string

func main() {
	// Check for arguments
	if len(os.Args) < 3 {
		fmt.Print("Did you provide enough arguments? -> " + os.Args[0] + " <username-to-find> <your-username> <your-password> \n")
		os.Exit(1)
	}

	usernameToFind = os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	liveChannels := getLiveChannels()
	startXMPP(username, password, liveChannels)

	for {
	}
}

func getLiveChannels() []string {
	// request and parse the front page
	resp, err := http.Get(baseURL + "livestreams/")
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.A && n.Parent != nil && n.Parent.Parent != nil {
			return scrape.Attr(n.Parent.Parent, "class") == "browse-main-videos--item"
		}
		return false
	}

	// grab all channels
	var liveChannels []string
	channels := scrape.FindAll(root, matcher)
	for _, channel := range channels {
		link := scrape.Attr(channel, "href")

		// links containing "videoes" are not live channels
		if !strings.Contains(link, "videos") {
			liveChannels = append(liveChannels, link[1:len(link)-1])
		}
	}

	return liveChannels
}

type SavedUser struct {
	Room     string
	Username string
}

var savedUsers []*SavedUser

func saveUser(from string) {
	splittedFrom := strings.Split(from, "@chat.livecoding.tv/")
	if len(splittedFrom) != 2 {
		return
	}

	s := &SavedUser{
		Room:     splittedFrom[0],
		Username: splittedFrom[1],
	}

	savedUsers = append(savedUsers, s)
}

func findUsername() {
	var rooms []string

	for _, u := range savedUsers {
		if u.Username == usernameToFind {
			rooms = append(rooms, u.Room)
		}
	}

	fmt.Print(usernameToFind + " was found in following channels: \n")
	for _, r := range rooms {
		fmt.Print(baseURL + r + "\n")
	}

	os.Exit(0)
}

func startXMPP(username, password string, liveChannels []string) {
	timeout := time.After(5 * time.Second)
	stopFetching := false

	options := xmpp.Options{
		Host:      "xmpp.livecoding.tv:5222",
		User:      username + jidSuffix,
		Password:  password,
		NoTLS:     true,
		Debug:     false,
		Session:   false,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client, err := options.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			if stopFetching {
				findUsername()
				return
			}
			chat, err := client.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Presence:
				saveUser(v.From)
			}
		}
	}()

	// Join all live channels to get users
	for _, channel := range liveChannels {
		client.JoinMUCNoHistory(channel+roomSuffix, username)
	}

	for {
		select {
		case <-timeout:
			stopFetching = true
			return
		}
		in := bufio.NewReader(os.Stdin)
		line, err := in.ReadString('\n')
		if err != nil {
			continue
		}
		line = strings.TrimRight(line, "\n")

		tokens := strings.SplitN(line, " ", 2)
		if len(tokens) == 2 {
			client.Send(xmpp.Chat{Remote: tokens[0], Type: "chat", Text: tokens[1]})
		}
	}
}
