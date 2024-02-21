package twitter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"github.com/dghubble/oauth1"
)

// Twitter representa o cliente para interagir com a API do Twitter
type Twitter struct {
	api *anaconda.TwitterApi // Cliente do Twitter
}

// NewTwitter cria uma nova instância do cliente do Twitter
func NewTwitter(groupID int64) (*Twitter, error) {
	promoGroupID := os.Getenv("PROMO_GROUP_ID")
	groupID, err := strconv.ParseInt(promoGroupID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Erro ao converter PROMO_GROUP_ID para int64: %v", err)
	}

	// Mapeamento explícito de PROMO_GROUP_ID para sufixo
	suffixMap := map[int64]string{
		-1002114057976: "1",
		2114057976:     "2",
		// Adicione mais mapeamentos conforme necessário
	}

	// Verificar se o PROMO_GROUP_ID tem um sufixo mapeado
	suffix, ok := suffixMap[groupID]
	if !ok {
		return nil, fmt.Errorf("Sufixo não mapeado para o PROMO_GROUP_ID: %d", groupID)
	}

	// Obter as credenciais do ambiente
	apiKeyEnv := fmt.Sprintf("TWITTER_API_KEY_%s", suffix)
	apiSecretEnv := fmt.Sprintf("TWITTER_API_SECRET_KEY_%s", suffix)
	accessTokenEnv := fmt.Sprintf("TWITTER_ACCESS_TOKEN_%s", suffix)
	accessSecretEnv := fmt.Sprintf("TWITTER_ACCESS_TOKEN_SECRET_%s", suffix)

	apiKey := os.Getenv(apiKeyEnv)
	apiSecret := os.Getenv(apiSecretEnv)
	accessToken := os.Getenv(accessTokenEnv)
	accessSecret := os.Getenv(accessSecretEnv)

	// Verificar se todas as credenciais foram fornecidas
	if apiKey == "" || apiSecret == "" || accessToken == "" || accessSecret == "" {
		return nil, fmt.Errorf("credenciais incompletas fornecidas para o grupo %d", groupID)
	}

	// Configurar a autenticação OAuth1 usando a nova função
	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessSecret, apiKey, apiSecret)

	return &Twitter{
		api: api,
	}, nil
}

func Post(groupID int64) {
	// Mapeamento explícito de PROMO_GROUP_ID para sufixo
	suffixMap := map[int64]string{
		-1002114057976: "1",
		2114057976:     "2",
		// Adicione mais mapeamentos conforme necessário
	}

	// Verificar se o PROMO_GROUP_ID tem um sufixo mapeado
	suffix := suffixMap[groupID]

	// Obter as credenciais do ambiente
	apiKeyEnv := fmt.Sprintf("TWITTER_API_KEY_%s", suffix)
	apiSecretEnv := fmt.Sprintf("TWITTER_API_SECRET_KEY_%s", suffix)
	accessTokenEnv := fmt.Sprintf("TWITTER_ACCESS_TOKEN_%s", suffix)
	accessSecretEnv := fmt.Sprintf("TWITTER_ACCESS_TOKEN_SECRET_%s", suffix)

	apiKey := os.Getenv(apiKeyEnv)
	apiSecret := os.Getenv(apiSecretEnv)
	accessToken := os.Getenv(accessTokenEnv)
	accessSecret := os.Getenv(accessSecretEnv)

	// Criar token OAuth1
	config := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Endpoint e dados a serem postados
	endpointURL := "https://api.twitter.com/2/tweets"
	data := map[string]string{"text": "Hello World!"}

	// Codificar os dados em JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Erro ao codificar os dados JSON:", err)
		return
	}

	// Criar requisição HTTP
	req, err := http.NewRequest("POST", endpointURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Erro ao criar requisição HTTP:", err)
		return
	}

	// Definir cabeçalhos da requisição
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "v2CreateTweetJS")

	// Executar a requisição HTTP
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição HTTP:", err)
		return
	}
	defer resp.Body.Close()

	// Verificar se a requisição foi bem-sucedida
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("Tweet postado com sucesso!")
	} else {
		fmt.Printf("Erro ao postar o tweet. Status code: %d %s\n", resp.StatusCode, resp.Status)
	}
}

// PostTweet publica um tweet no Twitter com mídia
func (t *Twitter) PostTweet(groupID int64, tweet string, mediaURL string, mediaType string) error {
	// Publicar o tweet com mídia
	_, err := t.api.PostTweet(tweet, nil)
	if err != nil {
		log.Printf("Erro ao postar tweet para o grupo %d: %v", groupID, err)
		return err
	}

	log.Printf("Tweet postado com sucesso para o grupo %d!", groupID)
	return nil
}
