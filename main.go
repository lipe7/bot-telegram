package main

import (
	"log"
	"os"
	"strconv"
	"sync"

	telegram "affiliate-ali-api/internal/telegram"

	"github.com/joho/godotenv"
)

var (
	activeBot *telegram.Bot
	botMutex  sync.Mutex
)

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

	// Verifica se já existe um bot ativo
	if activeBot == nil {
		// Se não houver um bot ativo, crie um novo
		botMutex.Lock()
		defer botMutex.Unlock()
		if activeBot == nil {
			bot, err := telegram.NewBot(botToken, groupID)
			if err != nil {
				log.Fatal(err)
			}
			activeBot = bot
		}
	}

	// Use o bot ativo
	activeBot.Run()
}
