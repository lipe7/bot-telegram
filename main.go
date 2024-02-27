package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	telegram "affiliate-ali-api/internal/telegram"

	"github.com/joho/godotenv"
)

var mutex = &sync.Mutex{}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env:", err)
	}

	// Obtém o token do bot do ambiente
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	// Obtenha o ID do grupo do ambiente
	groupIDStr := os.Getenv("BOT_GROUP_ID")

	// Converta o ID do grupo para int64
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		log.Fatal("ID do grupo do Telegram inválido:", err)
	}

	bot, err := telegram.NewBot(botToken, groupID)
	if err != nil {
		log.Fatal(err)
	}

	// Canal para sinalizar a interrupção
	interrupt := make(chan struct{})

	// Inicie o bot em uma gorrotina
	go func() {
		for {
			select {
			case <-interrupt:
				return
			default:
				// Executar o bot
				bot.Run()

				// Aguardar antes de tentar novamente
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Aguarde uma interrupção
	<-interrupt
}
