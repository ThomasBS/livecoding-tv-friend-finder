package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mattn/go-xmpp"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	// TODO: Check for arguments given to app
	if len(os.Args) < 2 {
		fmt.Print("You need to provide username and password \n")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	liveChannels := getLiveChannels()
	startXMPP(username, password, liveChannels)

	for {
	}
}

func getLiveChannels() []string {
	// request and parse the front page
	resp, err := http.Get("https://www.livecoding.tv/livestreams/")
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

var jidSuffix = "@livecoding.tv"
var roomSuffix = "@chat.livecoding.tv"
var client *xmpp.Client
var usersFetched bool

type SavedUser struct {
	Room     string
	Username string
}

var savedUsers []*SavedUser

func saveUser(from string) {
	splittedFrom := strings.Split(from, "@chat.livecoding.tv/")

	s := &SavedUser{
		Room:     splittedFrom[0],
		Username: splittedFrom[1],
	}

	savedUsers = append(savedUsers, s)
}

func startXMPP(username, password string, liveChannels []string) {
	// timeout := time.After(5 * time.Second)

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

	// TODO: Should not keep a persistent connection to the XMPP server.
	// Use a timer or something to recieved presence stanzas within X time and then quit.
	go func() {
		for {
			chat, err := client.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Presence:
				fmt.Print("presence received")
				saveUser(v.From)
			}
		}
	}()

	// Join all live channels to get users
	for _, channel := range liveChannels {
		client.JoinMUCNoHistory(channel+roomSuffix, "yaky")
	}

	// TODO: remove this infinite loop and use it in main.go (or just a select{})
	for {
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

func JoinRoom(room string) {
	client.JoinMUCNoHistory(room+roomSuffix, "yaky")
}
