package telegram

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"affiliate-ali-api/internal/common"
	"affiliate-ali-api/internal/twitter"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	botAPI  *tgbotapi.BotAPI
	groupID int64
}

func NewBot(botToken string, groupID int64) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	botAPI.Debug = true

	chatConfig := tgbotapi.ChatConfig{ChatID: groupID}
	chat, err := botAPI.GetChat(chatConfig)
	if err != nil {
		return nil, err
	}

	log.Printf("Bot iniciado como %s no grupo %s", botAPI.Self.UserName, chat.Title)

	return &Bot{
		botAPI:  botAPI,
		groupID: groupID,
	}, nil
}

func (b *Bot) Run() {
	serviceStartTime := time.Now()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil || update.Message.Time().Before(serviceStartTime) {
			continue
		}

		if update.Message.IsCommand() {
			handleCommand(b, update.Message)
			continue
		}

		if update.Message.Chat.ID == b.groupID {
			// Responder à mensagem
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Olá! Eu sou um bot e fui iniciado neste grupo.")
			_, err := b.botAPI.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else if update.Message.Chat.Type == "private" {
			handlePrivateMessage(b, update.Message)
		}
	}
}

func handleCommand(b *Bot, message *tgbotapi.Message) {
	previousGroup := os.Getenv("PROMO_GROUP_ID")
	commandWithPrefix := "-100" + message.CommandWithAt()
	os.Setenv("PROMO_GROUP_ID", commandWithPrefix)
	newGroup := os.Getenv("PROMO_GROUP_ID")

	newGroupINT, err := strconv.ParseInt(newGroup, 10, 64)
	if err != nil {
		log.Fatalf("Erro ao converter ID do grupo para inteiro: %v", err)
	}

	suffix, err := common.GetGroupSuffix(newGroupINT)
	if err != nil {
		log.Println("Erro ao obter sufixo:", err)
		return
	}

	groupNameEnv := fmt.Sprintf("GROUP_NAME_%s", suffix)
	groupName := os.Getenv(groupNameEnv)
	groupMsg := fmt.Sprintf("%s é o grupo configurado", groupName)

	msg := tgbotapi.NewMessage(message.Chat.ID, groupMsg)
	b.botAPI.Send(msg)

	log.Printf("\n______\nAtualizado:\nGrupo Anterior: %s\nNovo Grupo: %s\n______\n", previousGroup, newGroup)
}

func handlePrivateMessage(b *Bot, message *tgbotapi.Message) {
	forwardGroupIDStr := os.Getenv("PROMO_GROUP_ID")

	forwardGroupID, err := strconv.ParseInt(forwardGroupIDStr, 10, 64)
	if err != nil {
		log.Fatal("Invalid Telegram group ID:", err)
	}

	if hasImage(message) {
		photo := (*message.Photo)[len(*message.Photo)-1]
		photoID := photo.FileID
		caption := message.Caption

		forwardMsg := tgbotapi.NewPhotoShare(forwardGroupID, photoID)
		forwardMsg.Caption = caption

		_, err := b.botAPI.Send(forwardMsg)
		if err != nil {
			log.Println("Erro ao enviar imagem para o outro grupo:", err)
		}
	} else {
		textMsg := tgbotapi.NewMessage(forwardGroupID, message.Text)
		_, err := b.botAPI.Send(textMsg)
		if err != nil {
			log.Println("Erro ao postar mensagem de texto:", err)
		}
	}

	var tweetMessage string
	if message.Caption != "" {
		tweetMessage = message.Caption
	} else {
		tweetMessage = message.Text
	}

	promoGroupID := os.Getenv("PROMO_GROUP_ID")
	promoGroupIDInt, err := strconv.ParseInt(promoGroupID, 10, 64)
	if err != nil {
		log.Printf("Erro ao converter PROMO_GROUP_ID para int64: %v", err)
		return
	}

	twitter.Post(promoGroupIDInt, tweetMessage)
}

func hasImage(msg *tgbotapi.Message) bool {
	return msg.Photo != nil && len(*msg.Photo) > 0
}
