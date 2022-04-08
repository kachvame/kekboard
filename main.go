package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("`BOT_TOKEN` env variable is not set")
	}

	client, err := discordgo.New("Bot " + token)
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
