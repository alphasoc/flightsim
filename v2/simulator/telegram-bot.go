package simulator

import (
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
	URL                    = "https://api.telegram.org/bot"
	FlightsimTelegramToken = "FLIGHTSIM_TELEGRAM_TOKEN"
)

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
	return &TelegramBot{}
}

func printMsg(msg string) {
	fmt.Printf("%s [telegram-bot] %s\n", time.Now().Format("15:04:05"), msg)
}

// SendRequest sends a request to the Telegram Bot API,
// takes as an argument a bot method name and returns
// response status code.
// Bot methods list:
// https://core.telegram.org/bots/api#available-methods
func (tb *TelegramBot) SendRequest(botMethodName string) (int, error) {
	response, err := http.Get(tb.Url + botMethodName)
	if err == nil {
		response.Body.Close()
	}
	return response.StatusCode, err
}

func (tb *TelegramBot) GetMe() (int, error) {
	return tb.SendRequest("getMe")
}

func (tb *TelegramBot) GetUpdates() (int, error) {
	return tb.SendRequest("getUpdates")
}

func (tb *TelegramBot) GetMyCommands() (int, error) {
	return tb.SendRequest("getMyCommands")
}

// Simulate Telegram bot traffic
func (tb *TelegramBot) Simulate(ctx context.Context, host string) error {
	token := os.Getenv(FlightsimTelegramToken)

	if token == "" {
		printMsg("WARNING: No token in environment variable FLIGHTSIM_TELEGRAM_TOKEN was found. Using random string instead. This will generate traffic to api.telegram.org but return an authentication error. However, the traffic should still be captured by your SIEM.")
		token = generateRandomTelegramBotToken()
	}

	tb.Url = URL + token + "/"

	code, err := tb.GetMe()

	// If err is nil but return code is different than 200,
	// we return anyway -- there's no point in executing other commands
	// on server errors (most likely unauthorized due to random telegram token)
	if err != nil || code != 200 {
		return err
	}

	code, err = tb.GetMyCommands()
	if err != nil || code != 200 {
		return err
	}

	code, err = tb.GetUpdates()
	if err != nil || code != 200 {
		return err
	}

	return nil
}

// Init returns nil because TelegramBot module doesn't need a bind address.
func (TelegramBot) Init(bind BindAddr) error {
	return nil
}

func (TelegramBot) Cleanup() {

}

// Hosts returns a Telegram API domain name.
func (TelegramBot) Hosts(scope string, size int) ([]string, error) {
	return []string{"api.telegram.org"}, nil
}
