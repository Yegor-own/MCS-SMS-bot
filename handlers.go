package main

import (
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
)

// userStatuses:
// 		start
// 		set district
// 		set region
// 		get updates

func receivedMessageHandler(chatId string, userStatus string, receivedMessage string) {
	//fmt.Println("handler > rm ", receivedMessage, userStatus)
	chatId64, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic(err)
	}
	msg := tgbotapi.NewMessage(chatId64, "")
	if receivedMessage == "/resetRegion" {
		err := mc.Set(&memcache.Item{Key: strconv.Itoa(int(chatId64)), Value: NewSessionData("set district", "0", "0")})
		if err != nil {
			fmt.Println("handler > ", err)
		}

		text, buttons := getDistricts()
		msg.Text, msg.ReplyMarkup = "Выбери свой округ\n"+text, buttons
	}

	switch userStatus {
	case "start":
		err := mc.Set(&memcache.Item{Key: strconv.Itoa(int(chatId64)), Value: NewSessionData("set district", "0", "0")})
		if err != nil {
			fmt.Println("handler > ", err)
		}

		text, buttons := getDistricts()
		msg.Text, msg.ReplyMarkup = "Привет я СМС бот ⚙️, я буду присылать тебе уведомления о 🌦️ погодных 🌩️ условиях\nВыбери свой округ\n"+text, buttons
		fmt.Println("handler > ", "start")
	case "set district":
		fmt.Println("handler > ", "set district")
		i, err := strconv.Atoi(receivedMessage)
		if err != nil || i > 8 || i < 1 {
			fmt.Println("handler > ", err)
			msg.Text = "Введите номер округа"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}
		err = mc.Set(&memcache.Item{Key: strconv.Itoa(int(chatId64)), Value: NewSessionData("set region", receivedMessage, "0")})
		if err != nil {
			fmt.Println("handler > ", err)
			msg.Text = "Упс кажется что-то пошло не так"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}
		//fmt.Println("handler > rm ", receivedMessage)

		text, buttons := getRegions(receivedMessage)
		msg.Text, msg.ReplyMarkup = "Выберите ваш регион\n"+text, buttons
	case "set region":
		fmt.Println("handler > ", "set region")
		i, err := strconv.Atoi(receivedMessage)
		if err != nil || i < 1 || i > 111 {
			fmt.Println("handler > ", err)
			msg.Text = "Введи номер региона"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}

		it, err := mc.Get(strconv.Itoa(int(chatId64)))
		var data SessionData
		if err = json.Unmarshal(it.Value, &data); err != nil {
			fmt.Println("handler > ", err)
			msg.Text = "Упс кажется что-то пошло не так"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}

		err = mc.Set(&memcache.Item{Key: it.Key, Value: NewSessionData("get updates", data.DistrictId, receivedMessage)})
		if err != nil {
			fmt.Println("handler > ", err)
			msg.Text = "Упс кажется что-то пошло не так"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}

		text, err := getRegionById(receivedMessage, data.DistrictId)
		if err != nil {
			msg.Text = err.Error()
		} else {
			msg.Text = "Вы будете получать ежедневные оповещения о погодных условиях для " + text + "\nЧтобы получить оповещение сейчас нажмите /getAlert\nЧтобы сбросить выбраный регион используйте /resetRegion"
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("/getAlert"),
					tgbotapi.NewKeyboardButton("/resetRegion"),
				),
			)
		}

		at := AlertTimer{
			chatId64,
			receivedMessage,
		}
		go delayedAlert(at)
	case "get updates":
		fmt.Println("handler > ", "get updates")
		if receivedMessage == "/getAlert" {
			it, err := mc.Get(strconv.Itoa(int(chatId64)))
			if err != nil {
				fmt.Println("handler > ", err)
				msg.Text = "Упс кажется что-то пошло не так"
				if _, err = Bot.Send(msg); err != nil {
					fmt.Println("handler > ", err)
				}
				return
			}

			var data SessionData
			if err = json.Unmarshal(it.Value, &data); err != nil {
				fmt.Println("handler > ", err)
				msg.Text = "Упс кажется что-то пошло не так"
				if _, err = Bot.Send(msg); err != nil {
					fmt.Println("handler > ", err)
				}
				return
			}

			content := getAlertRegionsData(data.RegionId)

			var events = "Оповещения:"
			for _, event := range content.Events {
				events += "\n" + event
			}
			msg.Text = content.Region[1] + ":\n" + events
		}
	}

	if m, err := Bot.Send(msg); err != nil {
		fmt.Println(m, err)
	}
}
