package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dghubble/oauth1"
)

type Twitter struct {
	api *anaconda.TwitterApi
}

func NewTwitter(promoGroupID string) (*Twitter, error) {
	groupID, err := strconv.ParseInt(promoGroupID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter PROMO_GROUP_ID para int64: %v", err)
	}

	apiKey := os.Getenv("TWITTER_API_KEY")
	apiSecret := os.Getenv("TWITTER_API_SECRET_KEY")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	if apiKey == "" || apiSecret == "" || accessToken == "" || accessSecret == "" {
		return nil, fmt.Errorf("credenciais incompletas fornecidas para o grupo %d", groupID)
	}

	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessSecret, apiKey, apiSecret)

	return &Twitter{
		api: api,
	}, nil
}

func Post(groupID int64, tweetMessage string) {
	apiKey := os.Getenv("TWITTER_API_KEY")
	apiSecret := os.Getenv("TWITTER_API_SECRET_KEY")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	fmt.Print("\n")
	fmt.Print("\n apiKey", apiKey)
	fmt.Print("\n apiSecret", apiSecret)
	fmt.Print("\n accessToken", accessToken)
	fmt.Print("\n accessSecret", accessSecret)
	fmt.Print("\n")
	config := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	endpointURL := "https://api.twitter.com/2/tweets"
	data := map[string]string{"text": tweetMessage}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Erro ao codificar os dados JSON:", err)
		return
	}

	req, err := http.NewRequest("POST", endpointURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Erro ao criar requisição HTTP:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "v2CreateTweetJS")

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição HTTP:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("Tweet postado com sucesso!")
	} else {
		fmt.Printf("Erro ao postar o tweet. Status code: %d %s\n", resp.StatusCode, resp.Status)
	}
}
