package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/syndtr/goleveldb/leveldb"
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

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "kekboard.leveldb"
	}

	client, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("failed to create discord client:", err)
	}

	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		log.Fatalln("failed to open leveldb:", err)
	}

	defer func(db *leveldb.DB) {
		_ = db.Close()
	}(db)

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

	getMessageKey := func(reaction *discordgo.MessageReaction) []byte {
		return []byte(fmt.Sprintf("message-%s-%s", reaction.ChannelID, reaction.MessageID))
	}

	handleReaction := func(session *discordgo.Session, reaction *discordgo.MessageReaction) {
		message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
		if err != nil {
			log.Println("received reaction but unable to get message:", reaction.ChannelID, reaction.MessageID)

			return
		}

		messageKey := getMessageKey(reaction)
		kekCount := getKekCount(message)

		log.Printf("message %s has %d kek reactions", reaction.MessageID, kekCount)

		hasEnoughReactions := kekCount >= reactionThreshold

		_, err = db.Get(messageKey, nil)
		isOnKekboard := err == nil

		log.Println("has enough reactions", hasEnoughReactions, "is on kekboard", isOnKekboard)

		if hasEnoughReactions == isOnKekboard {
			// TODO: edit message if on kekboard

			return
		}

		// TODO: send to kekboard channel

		if hasEnoughReactions {
			log.Println("adding to kekboard")

			err = db.Put(messageKey, []byte("ok"), nil)
			if err != nil {
				log.Println("failed to add message to kekboard:", err)
			}

			return
		}

		log.Println("removing from kekboard")

		err = db.Delete(messageKey, nil)
		if err != nil {
			log.Println("failed to remove message from kekboard:", err)
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
