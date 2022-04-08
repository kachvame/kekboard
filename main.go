package main

import (
	"encoding/json"
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

type kekboardMessageState struct {
	ChannelID string
	MessageID string
	Reactions int
}

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

	kekboardChannelId := os.Getenv("KEKBOARD_CHANNEL_ID")
	if kekboardChannelId == "" {
		log.Fatalln("`KEKBOARD_CHANNEL_ID` env variable is not set")
	}

	kekEmoji := os.Getenv("KEK_EMOJI")
	if kekEmoji == "" {
		log.Fatalln("`KEK_EMOJI` env variable is not set")
	}

	client, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("failed to create discord client:", err)
	}

	client.Identify.Intents = discordgo.IntentsAll

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

	getMessageURL := func(message *discordgo.Message, guildID string) string {
		return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, message.ChannelID, message.ID)
	}

	getMessageEmbed := func(message *discordgo.Message, author *discordgo.Member, guildID string) []*discordgo.MessageEmbed {
		embeds := make([]*discordgo.MessageEmbed, 0, 4)
		embeds = append(embeds, &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    author.Nick,
				IconURL: author.AvatarURL("128"),
			},
			Description: message.Content,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Original",
					Value: fmt.Sprintf("[Jump!](%s)", getMessageURL(message, guildID)),
				},
			},
		})

		embedIdx := 0
		for _, attachment := range message.Attachments {
			if !strings.HasPrefix(attachment.ContentType, "image/") {
				continue
			}

			if embedIdx >= len(embeds) {
				embeds = append(embeds, &discordgo.MessageEmbed{})
			}

			embeds[embedIdx].URL = "https://twitter.com"
			embeds[embedIdx].Image = &discordgo.MessageEmbedImage{
				URL: attachment.URL,
			}

			embedIdx++
			if embedIdx >= 4 {
				break
			}
		}

		return embeds
	}

	putKekboardMessage := func(session *discordgo.Session, message *discordgo.Message, state *kekboardMessageState, guildID string) (*discordgo.Message, error) {
		guildMember, err := session.GuildMember(guildID, message.Author.ID)
		if err != nil {
			return nil, err
		}

		content := fmt.Sprintf("%s | %d", kekEmoji, state.Reactions)
		embeds := getMessageEmbed(message, guildMember, guildID)

		if state.MessageID == "" {
			return session.ChannelMessageSendComplex(state.ChannelID, &discordgo.MessageSend{
				Content: content,
				Embeds:  embeds,
			})
		}

		return session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			Channel: state.ChannelID,
			ID:      state.MessageID,
			Content: &content,
			Embeds:  embeds,
		})
	}

	handleReaction := func(session *discordgo.Session, reaction *discordgo.MessageReaction, guildID string) {
		message, err := session.ChannelMessage(reaction.ChannelID, reaction.MessageID)
		if err != nil {
			log.Println("received reaction but unable to get message:", reaction.ChannelID, reaction.MessageID)

			return
		}

		messageKey := getMessageKey(reaction)
		kekCount := getKekCount(message)

		hasEnoughReactions := kekCount >= reactionThreshold

		messageStateBytes, err := db.Get(messageKey, nil)

		var messageState *kekboardMessageState
		if err == nil {
			messageState = &kekboardMessageState{}
			if err = json.Unmarshal(messageStateBytes, messageState); err != nil {
				log.Println("failed to unmarshal state:", err)

				return
			}
		}

		isInKekboard := messageState != nil
		shouldEdit := isInKekboard && messageState.Reactions != kekCount

		if hasEnoughReactions && (!isInKekboard || shouldEdit) {
			if messageState == nil {
				messageState = &kekboardMessageState{
					ChannelID: kekboardChannelId,
				}
			}

			messageState.Reactions = kekCount

			message, err := putKekboardMessage(session, message, messageState, guildID)
			if err != nil {
				log.Println("failed to put message on kekboard:", err)

				return
			}

			messageState.MessageID = message.ID

			marshaledState, err := json.Marshal(messageState)
			if err != nil {
				log.Println("failed to marshal state:", err)

				return
			}

			err = db.Put(messageKey, marshaledState, nil)
			if err != nil {
				log.Println("failed to update state in leveldb:", err)
			}

			return
		}

		if isInKekboard && !hasEnoughReactions {
			err = session.ChannelMessageDelete(messageState.ChannelID, messageState.MessageID)
			if err != nil {
				log.Println("failed to delete message from kekboard:", err)
			}

			err = db.Delete(messageKey, nil)
			if err != nil {
				log.Println("failed to delete state from leveldb:", err)
			}
		}
	}

	client.AddHandler(func(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
		handleReaction(session, reactionAdd.MessageReaction, reactionAdd.GuildID)
	})

	client.AddHandler(func(session *discordgo.Session, reactionRemove *discordgo.MessageReactionRemove) {
		handleReaction(session, reactionRemove.MessageReaction, reactionRemove.GuildID)
	})

	client.AddHandler(func(session *discordgo.Session, reactionRemoveAll *discordgo.MessageReactionRemoveAll) {
		handleReaction(session, reactionRemoveAll.MessageReaction, reactionRemoveAll.GuildID)
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
