package simulator

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	URL                      = "https://api.telegram.org/bot"
	FLIGHTSIM_TELEGRAM_TOKEN = "FLIGHTSIM_TELEGRAM_TOKEN"
)

func readHTTPResponseBody(response *http.Response) (string, error) {
	var data string
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		data += scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return data, nil
}

func generateRandomTelegramBotToken() string {
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)

	// creating first part of token - 10 digits
	randomToken := ""
	for i := 0; i < 10; i++ {
		randomToken += strconv.Itoa(r.Intn(10))
	}

	randomToken += ":"

	// creating the rest of the token - 35 characters
	buffer := make([]byte, 35/2+1)
	_, _ = r.Read(buffer)

	randomToken += hex.EncodeToString(buffer)[:35]

	return randomToken
}

// TelegramBot simulates a Telegram bot traffic
type TelegramBot struct {
	Url string
}

// NewTelegramBot creates new TelegramBot simulator
func NewTelegramBot() *TelegramBot {
	token := os.Getenv(FLIGHTSIM_TELEGRAM_TOKEN)

	if "" == token {
		fmt.Println("WARNING: No token in environment variable FLIGHTSIM_TELEGRAM_TOKEN was found. Using random string instead.")
		token = generateRandomTelegramBotToken()
	}

	return &TelegramBot{Url: URL + token + "/"}
}

// SendRequest sends a request to the Telegram Bot API,
// takes as an argument a bot method name.
// Bot methods list:
// https://core.telegram.org/bots/api#available-methods
func (tb *TelegramBot) SendRequest(botMethodName string) (string, error) {
	response, err := http.Get(tb.Url + botMethodName)
	if err != nil {
		return "", err
	}
	return readHTTPResponseBody(response)
}

func (tb *TelegramBot) GetMe() (string, error) {
	return tb.SendRequest("getMe")
}

func (tb *TelegramBot) GetUpdates() (string, error) {
	return tb.SendRequest("getUpdates")
}

func (tb *TelegramBot) GetMyCommands() (string, error) {
	return tb.SendRequest("getMyCommands")
}

// Simulate Telegram bot traffic
func (tb *TelegramBot) Simulate(ctx context.Context, host string) error {
	_, err := tb.GetMe()
	if err != nil {
		return err
	}

	_, err = tb.GetMyCommands()
	if err != nil {
		return err
	}

	_, err = tb.GetUpdates()
	if err != nil {
		return err
	}

	return nil
}

// TelegramBot implements a Simulator interface but
// doesn't need a bind address
func (tb *TelegramBot) Init(bind BindAddr) error {
	return nil
}

func (tb *TelegramBot) Cleanup() {

}

// TelegramBot implements a Module interface but
// only host that is used is api.telegram.org
func (tb *TelegramBot) Hosts(scope string, size int) ([]string, error) {
	return []string{"api.telegram.org"}, nil
}
