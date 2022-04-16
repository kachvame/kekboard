package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type reactionContext struct {
	Session  *discordgo.Session
	Reaction *discordgo.MessageReaction
	GuildID  string
}

type kekboardMessageState struct {
	UserID    string
	GuildID   string
	ChannelID string
	MessageID string
	Reactions int
}

type statistics struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Count    int    `json:"count"`
}

var (
	statsCacheKey    = []byte("stats")
	messageKeyPrefix = []byte("message-")
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

	client.AddHandlerOnce(func(session *discordgo.Session, ready *discordgo.Ready) {
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
		return []byte(fmt.Sprintf("%s%s-%s", messageKeyPrefix, reaction.ChannelID, reaction.MessageID))
	}

	getMessageURL := func(message *discordgo.Message, guildID string) string {
		return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, message.ChannelID, message.ID)
	}

	getMessageEmbed := func(message *discordgo.Message, author *discordgo.Member, guildID string) []*discordgo.MessageEmbed {
		embeds := make([]*discordgo.MessageEmbed, 0, 4)

		authorName := author.Nick
		if authorName == "" {
			authorName = author.User.Username
		}

		embeds = append(embeds, &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name:    authorName,
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
		guildMember, err := session.State.Member(guildID, message.Author.ID)
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

	getMessage := func(session *discordgo.Session, channelID, messageID string) (*discordgo.Message, error) {
		cachedMessage, err := session.State.Message(channelID, messageID)
		if err == nil {
			return cachedMessage, nil
		}

		if err != discordgo.ErrStateNotFound {
			return nil, err
		}

		return session.ChannelMessage(channelID, messageID)
	}

	reactionCh := make(chan reactionContext, 128)
	done := make(chan bool, 1)

	defer func(done chan bool) {
		done <- true
	}(done)

	go func() {
		reactionHandler := func(ctx reactionContext) {
			message, err := getMessage(ctx.Session, ctx.Reaction.ChannelID, ctx.Reaction.MessageID)
			if err != nil {
				log.Println("received reaction but unable to get message:", ctx.Reaction.ChannelID, ctx.Reaction.MessageID)

				return
			}

			messageKey := getMessageKey(ctx.Reaction)
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
						UserID:    message.Author.ID,
						GuildID:   ctx.GuildID,
						ChannelID: kekboardChannelId,
					}
				}

				messageState.Reactions = kekCount

				message, err := putKekboardMessage(ctx.Session, message, messageState, ctx.GuildID)
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

				_ = db.Delete(statsCacheKey, nil)

				return
			}

			if isInKekboard && !hasEnoughReactions {
				err = ctx.Session.ChannelMessageDelete(messageState.ChannelID, messageState.MessageID)
				if err != nil {
					log.Println("failed to delete message from kekboard:", err)
				}

				err = db.Delete(messageKey, nil)
				if err != nil {
					log.Println("failed to delete state from leveldb:", err)
				}

				_ = db.Delete(statsCacheKey, nil)
			}

		}

		for ctx := range reactionCh {
			reactionHandler(ctx)
		}
	}()

	handleReaction := func(session *discordgo.Session, reaction *discordgo.MessageReaction, guildID string) {
		reactionCh <- reactionContext{
			Session:  session,
			Reaction: reaction,
			GuildID:  guildID,
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

	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
	}))

	router.Get("/stats", func(response http.ResponseWriter, request *http.Request) {
		cachedStatistics, err := db.Get(statsCacheKey, nil)
		if err == nil {
			_, _ = response.Write(cachedStatistics)

			return
		}

		type userKey struct {
			userID  string
			guildID string
		}

		userStatisticsMap := make(map[userKey]int)

		messageIterator := db.NewIterator(&util.Range{Start: messageKeyPrefix}, nil)
		for messageIterator.Next() {
			messageStateBytes := messageIterator.Value()

			messageState := &kekboardMessageState{}
			err := json.Unmarshal(messageStateBytes, messageState)
			if err != nil {
				continue
			}

			userStatisticsMap[userKey{messageState.UserID, messageState.GuildID}] += messageState.Reactions
		}
		messageIterator.Release()

		userStatistics := make([]statistics, 0, len(userStatisticsMap))
		for key, count := range userStatisticsMap {
			member, err := client.State.Member(key.guildID, key.userID)
			if err != nil {
				continue
			}

			user := member.User

			userStatistics = append(userStatistics, statistics{
				Username: user.Username,
				Avatar:   user.AvatarURL(""),
				Count:    count,
			})
		}

		sort.Slice(userStatistics, func(i, j int) bool {
			return userStatistics[i].Count > userStatistics[j].Count
		})

		statisticsBytes, _ := json.Marshal(userStatistics)

		_ = db.Put(statsCacheKey, statisticsBytes, nil)
		_, _ = response.Write(statisticsBytes)
	})

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	gracefulShutdown := func(server *http.Server, timeout time.Duration) error {
		done := make(chan error, 1)
		go func() {
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
			<-signalCh

			ctx := context.Background()

			var cancel context.CancelFunc
			if timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			done <- server.Shutdown(ctx)
		}()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}

		return <-done
	}

	if err := gracefulShutdown(httpServer, 10*time.Second); err != nil {
		log.Println("http server error:", err)
	}
}
