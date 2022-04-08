package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func main() {
	client, err := discordgo.New("TODO")
	if err != nil {
		log.Fatalln("failed to create discord client:", err)
	}

	if err = client.Open(); err != nil {
		log.Fatalln("failed to open discord session:", err)
	}

	defer func(client *discordgo.Session) {
		_ = client.Close()
	}(client)
}
