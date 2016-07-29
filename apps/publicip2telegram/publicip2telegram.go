package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/eternnoir/gotelebot"
)

type TelegramBot struct {
	bot *gotelebot.TeleBot
}

func main() {

	// Get params

	key := flag.String("key", "", "Your telegram key")
	device := flag.Int("chatId", 0, "The chat id")

	flag.Parse()

	// Create telegram bot

	bot := &TelegramBot{}
	bot.bot = gotelebot.InitTeleBot(*key)

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

	// Define current ip

	currentIp := ""

	// Get public ip and send to telegram if changed

	for {
		resp, err := http.Get("http://think-free.me/ip.php")
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Stderr.WriteString("\n")
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Stderr.WriteString("\n")
			continue
		}

		if currentIp != string(body) {

			currentIp = string(body)
			fmt.Println(currentIp)

			_, err := bot.bot.SendMessage(*device, "Ip changed : "+currentIp, nil)

			if err != nil {

				fmt.Println(err)
			}
		}

		time.Sleep(time.Minute)
	}
}
