package telegram

import (
	"fmt"
	"strconv"

	"github.com/eternnoir/gotelebot"
)

type TelegramBot struct {
	bot *gotelebot.TeleBot
}

func New(key string) *TelegramBot {

	bot := &TelegramBot{}
	bot.bot = gotelebot.InitTeleBot(key)

	newMsgChan := bot.bot.Messages

	go func() {

		for {

			select {
			case m := <-newMsgChan:

				if m.Text != "" {
					fmt.Println("Got message from " + strconv.FormatInt(int64(m.Chat.Id), 10) + " : " + m.Text)
					bot.bot.SendMessage(int(m.Chat.Id), "Received : "+m.Text, nil)
				}
			}
		}
	}()

	return bot
}

func (bot *TelegramBot) Send(device string, message string) {

	i, _ := strconv.Atoi(device)
	bot.bot.SendMessage(i, message, nil)
}
