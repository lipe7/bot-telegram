package telegram

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Bot representa o bot do Telegram
type Bot struct {
	botAPI  *tgbotapi.BotAPI
	groupID int64
}

// NewBot cria uma nova instância do bot do Telegram
func NewBot(botToken string, groupID int64) (*Bot, error) {
	// Crie uma nova instância do bot
	botAPI, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	// Se deseja ver a interação no log
	botAPI.Debug = true

	// Obter informações sobre o grupo
	chatConfig := tgbotapi.ChatConfig{ChatID: groupID}
	log.Println("chatConfig: ", groupID)

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

// Run inicia o bot e começa a lidar com as mensagens
func (b *Bot) Run() {
	// Configurar uma atualização para pegar mensagens
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	// Loop pelas mensagens recebidas
	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Verificar se a mensagem veio do grupo especificado
		if update.Message.Chat.ID == b.groupID {
			// Responder à mensagem
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Olá! Eu sou um bot e fui iniciado neste grupo.")
			_, err := b.botAPI.Send(msg)
			if err != nil {
				log.Println(err)
			}
		} else if update.Message.Chat.Type == "private" {
			// Reencaminhar a mensagem para o grupo
			forwardGroupIDStr := os.Getenv("TECH_PROMO_GROUP_ID")

			// Converta o ID do grupo para int64
			forwardGroupID, err := strconv.ParseInt(forwardGroupIDStr, 10, 64)
			if err != nil {
				log.Fatal("ID do grupo do Telegram inválido:", err)
			}

			forwardMsg := tgbotapi.ForwardConfig{BaseChat: tgbotapi.BaseChat{ChatID: forwardGroupID}, FromChatID: update.Message.Chat.ID, MessageID: update.Message.MessageID}
			_, err = b.botAPI.Send(forwardMsg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
