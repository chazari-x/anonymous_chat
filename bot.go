package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/m/v2/model"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var botMessage model.BotMessage
var err error

type conf struct {
	// Телеграмм бот
	Bot struct {
		Token string `yaml:"token"`
		URL   string `yaml:"url"`
	} `yaml:"bot"`
}

type Bot struct {
	bot  *tgbotapi.BotAPI
	conf *conf
}

func NewBot(config *conf) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(config.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("new bot api err: %s", err.Error())
	}

	return &Bot{bot: bot, conf: config}, nil
}

func (b *Bot) StartBot() error {
	url := fmt.Sprintf("%sbot%s", b.conf.Bot.URL, b.conf.Bot.Token)
	offset := 0

	go func() {
		if err = b.StartRoom(); err != nil {
			log.Println(err)
		}
	}()

	go model.RoomWriter()

	for {
		if updates, err := getUpdate(url, offset); err != nil {
			if strings.Contains(err.Error(), "connection reset by peer") {
				log.Printf("get update err: connection reset by peer")
			} else {
				return fmt.Errorf("get update err: %s", err.Error())
			}
		} else {
			for _, update := range updates {
				go func() {
					if err := b.startSendButtons(update); err != nil {
						log.Printf("start send buttons err: %s", err.Error())
					}
				}()

				offset = update.UpdateId + 1
			}
		}
	}
}

func (b *Bot) StartRoom() error {
	for {
		oneID, twoID := model.GetFromWaitingList()
		if oneID != 0 && twoID != 0 {
			model.AddToRoom(oneID, twoID)
			fmt.Println("---------------")
			fmt.Println("Изменение комнаты", model.R)
			fmt.Println("---------------")
			time.Sleep(time.Millisecond * 100)
			if err := b.sendButtons(oneID, twoID, "Собеседник найден."+
				"\n\nВведите команду /next для поиска нового собеседника "+
				"или нажмите кнопку \"🔍 Найти нового собеседника\""+
				"\n\nВведите команду /stop для прекращения диалога "+
				"или нажмите кнопку \"⛔ Закончить диалог\""); err != nil {
				return err
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func getUpdate(url string, offset int) ([]model.Update, error) {
	resp, err := http.Get(url + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, fmt.Errorf("http get err: %s", err.Error())
	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			log.Fatalln(err.Error())
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read all body err: %s", err.Error())
	}

	var restResponse model.RestResponse
	if json.Unmarshal(body, &restResponse) != nil {
		return nil, fmt.Errorf("json unmarshal err: %s", err.Error())
	}

	return restResponse.Result, nil
}

func (b *Bot) startSendButtons(update model.Update) error {
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = update.Message.Text

	if err = b.buttons(botMessage.ChatId, botMessage.Text); err != nil {
		msg := tgbotapi.NewMessage(botMessage.ChatId, "Произошла ошибка на сервере!")
		if _, err2 := b.bot.Send(msg); err2 != nil {
			return fmt.Errorf(err.Error() + "\n" + err2.Error())
		}

		return err
	}

	return nil
}

func (b *Bot) buttons(oneID int64, text string) error {
	var twoID int64
	index := model.GetUser(botMessage.ChatId).Index

	switch index {
	case "start_chat", "restart_chat":
		switch text {
		case "⛔ Остановить поиск собеседника", "/stop":
			model.DeleteFromWaitingList(oneID)
			index = "home"
			text = "Поиск собеседника остановлен. " +
				"\n\nВведите команду /find для поиска собеседника " +
				"или нажмите кнопку \"🔍 Найти собеседника\"."
			model.UpdateUser(oneID, index)
			if err := b.sendButtons(oneID, twoID, text); err != nil {
				return err
			}
		}

	case "chatting":
		switch text {
		case "🔍 Найти другого собеседника", "/next", "\"🔍 Найти собеседника\"":
			twoID = model.RestartRoom(oneID)
			index = "restart_chat"
			text = "Идет поиск другого собеседника..." +
				"\n\nВведите команду /stop для прекращения поиска " +
				"или нажмите кнопку \"⛔ Остановить поиск собеседника\""
			model.UpdateUser(oneID, index)
			if err := b.sendButtons(oneID, twoID, text); err != nil {
				return err
			}
		case "⛔ Закончить диалог", "/stop", "⛔ Остановить поиск собеседника":
			twoID = model.DeleteRoom(oneID)
			index = "home"
			text = "Диалог закончен. " +
				"\n\nВведите команду /find для поиска собеседника " +
				"или нажмите кнопку \"🔍 Найти собеседника\"."
			model.UpdateUser(oneID, index)
			if err := b.sendButtons(oneID, twoID, text); err != nil {
				return err
			}
		default:
			index = "message"
			if err := b.sendMessage(oneID, text); err != nil {
				return err
			}
		}

	default:
		switch text {
		case "🔍 Найти собеседника", "🔍 Найти другого собеседника", "/next", "/find":
			index = "start_chat"
			text = "Идет поиск собеседника..." +
				"\n\nВведите команду /stop для прекращения поиска " +
				"или нажмите кнопку \"⛔ Остановить поиск собеседника\""

		default:
			index = "home"
			text = "Введите команду /find для поиска собеседника " +
				"или нажмите кнопку \"🔍 Найти собеседника\"."
		}
		model.UpdateUser(oneID, index)
		if err := b.sendButtons(oneID, twoID, text); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bot) sendButtons(oneID, twoID int64, text string) error {
	var keyboard tgbotapi.ReplyKeyboardMarkup
	var row []tgbotapi.KeyboardButton
	var buttons []string
	var index = model.GetUser(botMessage.ChatId).Index

	switch index {
	case "home":
		buttons = []string{"🔍 Найти собеседника"}

	case "start_chat", "restart_chat":
		model.AddToWaitingList(oneID)
		switch index {
		case "start_chat":
			buttons = append(buttons, "⛔ Остановить поиск собеседника")
		case "restart_chat":
			buttons = append(buttons, "⛔ Остановить поиск собеседника")
		}
	case "chatting":
		buttons = append(buttons, "🔍 Найти другого собеседника", "⛔ Закончить диалог")
	}

	for i := range buttons {
		btn := tgbotapi.NewKeyboardButton(buttons[i])
		row := append(row, btn)
		keyboard.Keyboard = append(keyboard.Keyboard, row)
	}

	fmt.Println("---------------")
	fmt.Println("Пользователи", model.U)
	fmt.Println("Лист ожидания", model.W)
	fmt.Println("Комнаты", model.R)
	fmt.Println("---------------")

	keyboard.ResizeKeyboard = true

	switch index {
	case "start_chat":
		message := tgbotapi.NewMessage(oneID, text)
		message.ReplyMarkup = keyboard
		if _, err = b.bot.Send(message); err != nil {
			return err
		}

	case "home", "restart_chat":
		if text == "Поиск собеседника остановлен. "+
			"\n\nВведите команду /find для поиска собеседника "+
			"или нажмите кнопку \"🔍 Найти собеседника\"." ||
			text == "Введите команду /find для поиска собеседника "+
				"или нажмите кнопку \"🔍 Найти собеседника\"." {
			message := tgbotapi.NewMessage(oneID, text)
			message.ReplyMarkup = keyboard
			if _, err = b.bot.Send(message); err != nil {
				return err
			}
		} else {
			message := tgbotapi.NewMessage(oneID, text)
			message.ReplyMarkup = keyboard
			if _, err = b.bot.Send(message); err != nil {
				return err
			}

			var row1 []tgbotapi.KeyboardButton
			var key tgbotapi.ReplyKeyboardMarkup
			key.ResizeKeyboard = true
			row1 = append(row1, tgbotapi.NewKeyboardButton("⛔ Остановить поиск собеседника"))
			key.Keyboard = append(key.Keyboard, row1)
			model.UpdateUser(twoID, "restart_chat")
			msg := tgbotapi.NewMessage(twoID, "Собеседник ушёл. Поиск нового собеседника."+
				"\n\nВведите команду /stop для прекращения поиска "+
				"или нажмите кнопку \"⛔ Остановить поиск собеседника\"")
			msg.ReplyMarkup = key
			if _, err = b.bot.Send(msg); err != nil {
				return err
			}
		}

	case "chatting":
		message := tgbotapi.NewMessage(oneID, text)
		message.ReplyMarkup = keyboard
		if _, err = b.bot.Send(message); err != nil {
			return err
		}

		message = tgbotapi.NewMessage(twoID, text)
		message.ReplyMarkup = keyboard
		if _, err = b.bot.Send(message); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bot) sendMessage(oneID int64, text string) error {
	var twoID = model.RoomUser(oneID)

	message := tgbotapi.NewMessage(twoID, text)
	message.ParseMode = "Markdown"
	if _, err = b.bot.Send(message); err != nil {
		return fmt.Errorf(strconv.FormatInt(twoID, 10), err.Error())
	}

	return nil
}
