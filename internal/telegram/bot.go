package telegram

import (
	"log"
	"os"
	"strconv"
	"time"

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
	serviceStartTime := time.Now()

	// Configurar uma atualização para pegar mensagens
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botAPI.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	// Loop pelas mensagens recebidas
	for update := range updates {
		if update.Message == nil || update.Message.Time().Before(serviceStartTime) {
			continue
		}

		// Verificar se a mensagem é um comando
		if update.Message.IsCommand() {
			previousGroup := os.Getenv("PROMO_GROUP_ID") // Obter o valor anterior do env

			// Atribuir o novo valor com o prefixo ao env PROMO_GROUP_ID
			commandWithPrefix := "-100" + update.Message.CommandWithAt()
			os.Setenv("PROMO_GROUP_ID", commandWithPrefix)

			newGroup := os.Getenv("PROMO_GROUP_ID") // Obter o novo valor do env após atualizar

			// Imprimir o log conforme especificado
			log.Printf("\n______\nAtualizado:\nGrupo Anterior: %s\nNovo Grupo: %s\n______\n", previousGroup, newGroup)
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
			forwardGroupIDStr := os.Getenv("PROMO_GROUP_ID")

			// Converta o ID do grupo para int64
			forwardGroupID, err := strconv.ParseInt(forwardGroupIDStr, 10, 64)
			if err != nil {
				log.Fatal("Invalid Telegram group ID:", err)
			}

			forwardMsg := tgbotapi.ForwardConfig{
				BaseChat:   tgbotapi.BaseChat{ChatID: forwardGroupID},
				FromChatID: update.Message.Chat.ID,
				MessageID:  update.Message.MessageID,
			}
			_, err = b.botAPI.Send(forwardMsg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
