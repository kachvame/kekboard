package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	_ = godotenv.Load()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("`BOT_TOKEN` env variable is not set")
	}

	reactionThresholdStr := os.Getenv("REACTION_THRESHOLD")
	reactionThreshold, err := strconv.Atoi(reactionThresholdStr)
	if err != nil {
		log.Fatalln("failed to parse `REACTION_THRESHOLD` env variable:", err)
	}

	emojiTarget := os.Getenv("EMOJI_TARGET")
	if emojiTarget == "" {
		emojiTarget = "kek"
	}

	client, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("failed to create discord client:", err)
	}

	client.AddHandlerOnce(func(_ *discordgo.Session, ready *discordgo.Ready) {
		log.Printf("Logged in as %s#%s\n", ready.User.Username, ready.User.Discriminator)
	})

	getKekCount := func(message *discordgo.Message) (count int) {
		for _, reaction := range message.Reactions {
			if strings.Contains(strings.ToLower(reaction.Emoji.Name), emojiTarget) {
				count += reaction.Count
			}
		}

		return
	}

	handleReaction := func(session *discordgo.Session, reaction *discordgo.MessageReaction) {
		message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
		if err != nil {
			log.Println("received reaction but unable to get message:", reaction.ChannelID, reaction.MessageID)

			return
		}

		kekCount := getKekCount(message)
		log.Printf("message %s has %d kek reactions", reaction.MessageID, kekCount)

		if kekCount >= reactionThreshold {
			log.Println("message has enough reactions:", reaction.MessageID)
		}
	}

	client.AddHandler(func(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
		handleReaction(session, reactionAdd.MessageReaction)
	})

	client.AddHandler(func(session *discordgo.Session, reactionRemove *discordgo.MessageReactionRemove) {
		handleReaction(session, reactionRemove.MessageReaction)
	})

	client.AddHandler(func(session *discordgo.Session, reactionRemoveAll *discordgo.MessageReactionRemoveAll) {
		handleReaction(session, reactionRemoveAll.MessageReaction)
	})

	if err = client.Open(); err != nil {
		log.Fatalln("failed to open discord session:", err)
	}

	defer func(client *discordgo.Session) {
		_ = client.Close()
	}(client)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-signalCh
}
