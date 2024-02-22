package telegram

import (
	"affiliate-ali-api/internal/twitter"
	"fmt"
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

	// Criar uma nova instância do cliente do Twitter
	promoGroupID := os.Getenv("PROMO_GROUP_ID")
	groupID, err := strconv.ParseInt(promoGroupID, 10, 64)

	// Verificar se houve um erro ao converter promoGroupID para int64
	if err != nil {
		log.Printf("Erro ao converter PROMO_GROUP_ID para int64: %v", err)
		return
	}

	// Criar o cliente do Twitter com o groupID
	twitter.NewTwitter(groupID)

	// Verificar se houve um erro ao criar o cliente do Twitter
	if err != nil {
		log.Printf("Erro ao criar cliente do Twitter para o grupo %d: %v", groupID, err)
		return
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
			// Verificar se a mensagem contém uma foto
			// if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
			// Se houver uma foto, obter a maior resolução disponível
			// photo := (*update.Message.Photo)[len(*update.Message.Photo)-1] // Pegue a última foto, que é a maior
			// photoURL := photo.FileID

			// Construir a mensagem do tweet com a descrição da foto, se houver
			// var tweetMessage string

			// if update.Message.Caption != "" {
			// 	tweetMessage = update.Message.Caption + "\n" + photoURL
			// } else {
			// 	tweetMessage = photoURL // Corrigido: Se não houver legenda, defina a mensagem como a URL da foto
			// }

			// Postar o tweet no Twitter
			// err := twitterClient.PostTweet(b.groupID, tweetMessage, "tweetPhoto", "")
			// if err != nil {
			// 	log.Println("Erro ao postar no Twitter:", err)
			// }
			// } else {
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
			} else {

				twitter.Post(promoGroupIDInt, tweetMessage)
			}
			// }

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
