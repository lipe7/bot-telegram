package telegram

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"affiliate-ali-api/internal/twitter"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	botAPI  *tgbotapi.BotAPI
	groupID int64
}

var activeBots map[string]*Bot = make(map[string]*Bot)

func NewBot(botToken string, groupID int64) (*Bot, error) {
	// Verificar se já existe uma instância ativa com o mesmo botToken
	if existingBot, ok := activeBots[botToken]; ok {
		log.Printf("Utilizando instância existente do bot com o token %s", botToken)
		return existingBot, nil
	}

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

	// Criar nova instância do bot e adicionar ao mapa de bots ativos
	newBot := &Bot{
		botAPI:  botAPI,
		groupID: groupID,
	}
	activeBots[botToken] = newBot

	return newBot, nil
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
		if update.Message.IsCommand() {
			handleCommand(b, update.Message)
			continue
		}

		if !isAuthenticated(update.Message.From.ID) {
			authenticateUser(b, update.Message.From.ID, update.Message.Text)
			continue
		}

		if update.Message == nil || update.Message.Time().Before(serviceStartTime) {
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
	if message.Command() == "start" {
		// Verifica se o usuário já está autenticado
		if !isAuthenticated(message.From.ID) {
			// Solicita ao usuário que insira o código de autenticação
			userIDStr := strconv.Itoa(message.From.ID)
			userIDInt, err := strconv.ParseInt(userIDStr, 10, 64)
			if err != nil {
				log.Printf("Erro ao converter userID para int64: %v", err)
				return
			}

			authMsg := tgbotapi.NewMessage(userIDInt, "Por favor, insira o código de autenticação:")
			_, err = b.botAPI.Send(authMsg)
			if err != nil {
				log.Printf("Erro ao enviar mensagem de autenticação: %v", err)
			}
			return
		}
	}
}

func handlePrivateMessage(b *Bot, message *tgbotapi.Message) {
	if !isAuthenticated(message.From.ID) {
		userIDStr := strconv.Itoa(message.From.ID)
		userIDInt, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			log.Printf("Erro ao converter userID para int64: %v", err)
			return
		}

		// Usuário não autenticado, ignorar a mensagem
		authMsg := tgbotapi.NewMessage(userIDInt, "Usuário não autenticado, por favor, insira o código de autenticação:")
		_, err = b.botAPI.Send(authMsg)
		if err != nil {
			log.Printf("Erro ao enviar mensagem de autenticação: %v", err)
		}
		return
	}

	forwardGroupIDStr := os.Getenv("GROUP_ID")

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

	promoGroupID := os.Getenv("GROUP_ID")
	promoGroupIDInt, err := strconv.ParseInt(promoGroupID, 10, 64)
	if err != nil {
		log.Printf("Erro ao converter GROUP_ID para int64: %v", err)
		return
	}

	twitter.Post(promoGroupIDInt, tweetMessage)
}

func hasImage(msg *tgbotapi.Message) bool {
	return msg.Photo != nil && len(*msg.Photo) > 0
}

func isAuthenticated(userID int) bool {
	// Verificar se o userID está na lista de usuários autenticados
	authenticatedUsers := os.Getenv("AUTHENTICATED_USERS")
	return strings.Contains(authenticatedUsers, strconv.Itoa(userID))
}

func authenticateUser(b *Bot, userID int, authCode string) {
	userIDStr := strconv.Itoa(userID)
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Erro ao converter userID para int64: %v", err)
		return
	}

	envAuthCode := os.Getenv("AUTH_CODE")
	envAuthCode = strings.TrimSpace(envAuthCode)
	if authCode != envAuthCode {
		// Código de autenticação incorreto, enviar mensagem informando
		errorMsg := "Código de autenticação incorreto. Por favor, tente novamente."
		errMsg := tgbotapi.NewMessage(userIDInt, errorMsg)
		_, err := b.botAPI.Send(errMsg)
		if err != nil {
			log.Printf("Erro ao enviar mensagem de erro de autenticação: %v", err)
		}
		return
	}

	// Adicionar o ID do usuário à lista de usuários autenticados
	authenticatedUsers := os.Getenv("AUTHENTICATED_USERS")
	authenticatedUsers += strconv.Itoa(userID) + ","
	os.Setenv("AUTHENTICATED_USERS", authenticatedUsers)

	// Solicitar ao usuário que insira o código de autenticação
	authMsg := tgbotapi.NewMessage(userIDInt, "Autenticado com sucesso.")
	_, err = b.botAPI.Send(authMsg)
	if err != nil {
		log.Printf("Erro ao enviar mensagem de autenticação: %v", err)
	}
}
