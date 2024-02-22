package telegram

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"affiliate-ali-api/internal/twitter"

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

			suffixMap := map[int64]string{
				-1002114057976: "1",
				-1002073907096: "2",
				// Adicione mais mapeamentos conforme necessário
			}

			newGroupINT, err := strconv.ParseInt(newGroup, 10, 64)
			if err != nil {
				log.Fatalf("Erro ao converter ID do grupo para inteiro: %v", err)
			}

			suffix := suffixMap[newGroupINT]
			groupNameEnv := fmt.Sprintf("GROUP_NAME_%s", suffix)
			groupName := os.Getenv(groupNameEnv)
			groupMsg := fmt.Sprintf("%s é o grupo configurado", groupName)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, groupMsg)
			b.botAPI.Send(msg)

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

			// Verificar se a mensagem contém uma imagem
			if hasImage(update.Message) {
				// Se houver uma imagem, obter a maior resolução disponível
				photo := (*update.Message.Photo)[len(*update.Message.Photo)-1] // Pegue a última foto, que é a maior
				photoID := photo.FileID
				caption := update.Message.Caption

				// Criar a configuração para enviar a foto e a legenda para o outro grupo
				forwardMsg := tgbotapi.NewPhotoShare(forwardGroupID, photoID)
				forwardMsg.Caption = caption

				// Enviar a foto e a legenda para o outro grupo
				_, err := b.botAPI.Send(forwardMsg)
				if err != nil {
					log.Println("Erro ao enviar imagem para o outro grupo:", err)
				}
			} else {
				// Se não houver imagem, postar o texto da mensagem
				textMsg := tgbotapi.NewMessage(forwardGroupID, update.Message.Text)
				_, err := b.botAPI.Send(textMsg)
				if err != nil {
					log.Println("Erro ao postar mensagem de texto:", err)
				}
			}

			// Postar a mensagem no Twitter
			var tweetMessage string
			if update.Message.Caption != "" {
				tweetMessage = update.Message.Caption
			} else {
				tweetMessage = update.Message.Text
			}

			promoGroupID := os.Getenv("PROMO_GROUP_ID")
			promoGroupIDInt, err := strconv.ParseInt(promoGroupID, 10, 64)
			if err != nil {
				log.Printf("Erro ao converter PROMO_GROUP_ID para int64: %v", err)
				return
			}

			twitter.Post(promoGroupIDInt, tweetMessage)
		}
	}
}

// Função auxiliar para verificar se a mensagem contém uma imagem
func hasImage(msg *tgbotapi.Message) bool {
	return msg.Photo != nil && len(*msg.Photo) > 0
}
