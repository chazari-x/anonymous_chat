package main

import (
	"encoding/json"
	"example.com/m/v2/model"
	"fmt"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
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
			return fmt.Errorf("get update err: %s", err.Error())
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
			fmt.Println("Изменение комнаты", model.R)

			if err := b.sendButtons(oneID, twoID, "Собеседник найден"); err != nil {
				return err
			}
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func getUpdate(url string, offset int) ([]model.Update, error) {
	resp, err := http.Get(url + "/getUpdates" + "?offset=" + strconv.Itoa(offset))
	if err != nil {
		return nil, fmt.Errorf("http get err: %s", err.Error())
	}

	defer resp.Body.Close()
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
		case "Остановить поиск собеседника", "/stop":
			model.DeleteFromWaitingList(oneID)
			index = "home"
			text = "Поиск собеседника остановлен. Нажмите кнопку \"Найти собеседника\"."
		}

		model.UpdateUser(oneID, index)

	case "chatting":
		switch text {
		case "Найти другого собеседника", "/next":
			twoID = model.RestartRoom(oneID)
			index = "restart_chat"
			text = "Идет поиск другого собеседника..."
			model.UpdateUser(oneID, index)

		case "Закончить диалог", "/stop":
			twoID = model.DeleteRoom(oneID)
			index = "home"
			text = "Диалог закончен. Нажмите кнопку \"Найти собеседника\"."
			model.UpdateUser(oneID, index)

		default:
			index = "message"

		}

	default:
		switch text {
		case "Найти собеседника", "Найти другого собеседника", "/next":
			index = "start_chat"
			text = "Идет поиск собеседника..."

		case "Остановить поиск собеседника", "/stop", "Закончить диалог":
			index = "home"
			text = "Нажмите кнопку \"Найти собеседника\"."
		}

		model.UpdateUser(oneID, index)
	}

	if index != "message" {
		if err := b.sendButtons(oneID, twoID, text); err != nil {
			return err
		}
	} else {
		if err := b.sendMessage(oneID, text); err != nil {
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
		buttons = []string{"Найти собеседника"}

	case "start_chat", "restart_chat":
		model.AddToWaitingList(oneID)
		switch index {
		case "start_chat":
			buttons = append(buttons, "Остановить поиск собеседника")
		case "restart_chat":
			buttons = append(buttons, "Остановить поиск собеседника")
		}
	case "chatting":
		buttons = append(buttons, "Найти другого собеседника", "Закончить диалог")
	}

	for i := range buttons {
		btn := tgbotapi.NewKeyboardButton(buttons[i])
		row := append(row, btn)
		keyboard.Keyboard = append(keyboard.Keyboard, row)
	}

	fmt.Println("Пользователи", model.U)
	fmt.Println("Лист ожидания", model.W)
	fmt.Println("Комнаты", model.R)

	keyboard.ResizeKeyboard = true

	switch index {
	case "start_chat":
		message := tgbotapi.NewMessage(oneID, text)
		message.ReplyMarkup = keyboard
		if _, err = b.bot.Send(message); err != nil {
			return err
		}

	case "home", "restart_chat":
		if text == "Поиск собеседника остановлен. Нажмите кнопку \"Найти собеседника\"." ||
			text == "Нажмите кнопку \"Найти собеседника\"." {
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
			row1 = append(row1, tgbotapi.NewKeyboardButton("Остановить поиск собеседника"))
			key.Keyboard = append(key.Keyboard, row1)
			model.UpdateUser(twoID, "restart_chat")
			msg := tgbotapi.NewMessage(twoID, "Собеседник ушёл. Поиск нового собеседника.")
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
