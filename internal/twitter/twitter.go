package twitter

import (
	"log"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

// PostTweet publica um tweet no Twitter usando a chave correta da API associada à conta
func PostTweet(text, twitterAccount string) error {
	apiKey := os.Getenv(twitterAccount + "_API_KEY")
	apiSecret := os.Getenv(twitterAccount + "_API_SECRET")
	accessToken := os.Getenv(twitterAccount + "_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv(twitterAccount + "_ACCESS_TOKEN_SECRET")

	config := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Cliente do Twitter
	client := twitter.NewClient(httpClient)

	// Parâmetros do tweet
	params := &twitter.StatusUpdateParams{
		Status: text,
	}

	// Publicar tweet
	_, _, err := client.Statuses.Update(text, params)
	if err != nil {
		log.Println("Erro ao postar tweet:", err)
		return err
	}

	return nil
}
