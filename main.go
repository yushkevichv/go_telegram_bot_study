package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const baseTelegramUrl = "https://api.telegram.org"
const telegramToken = "1397672646:AAEQDlDkCfyQYHGMjZHo7N4s8Zs2TT0QQEc"
const getUpdatesUri = "getUpdates"
const sendMessageUri = "sendMessage"

const keywordStart = "/start"

type UpdateT struct {
	Ok     bool            `json:"ok"`
	Result []UpdateResultT `json:"result"`
}

type UpdateResultT struct {
	UpdateId int                  `json:"update_id"`
	Message  UpdateResultMessageT `json:"message"`
}

type UpdateResultMessageT struct {
	MessageId int                     `json:"message_id"`
	From      UpdateResultFromT       `json:"from"`
	Chat      UpdateResultChatT       `json:"chat"`
	Date      int                     `json:"date"`
	Text      string                  `json:"text"`
	Entities  []UpdateResultEntitiesT `json:"entities,omitempty"`
}

type UpdateResultFromT struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Language  string `json:"language_code"`
}

type UpdateResultChatT struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

type UpdateResultEntitiesT struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
}

type SendMessageResponseT struct {
	Ok     bool                 `json:"ok"`
	Result UpdateResultMessageT `json:"result"`
}

func main() {
	ticker := time.NewTicker(5 * time.Second)
	LastUpdateId := 0
	anonymPairs := make(map[int]int)
	var tmpPairKey int

	for _ = range ticker.C {
		fmt.Println("last updateId: ", LastUpdateId)

		update, err := getUpdates(LastUpdateId)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for _, item := range update.Result {

			if tmpPairKey == 0 {
				tmpPairKey = item.Message.From.Id
			} else if tmpPairKey != item.Message.From.Id {
				anonymPairs[tmpPairKey] = item.Message.From.Id
				anonymPairs[item.Message.From.Id] = tmpPairKey
			}

			LastUpdateId = item.UpdateId

			if strings.Contains(strings.ToLower(item.Message.Text), "#anonym") {
				sendTo := anonymPairs[item.Message.From.Id]
				if sendTo != 0 {
					result, err := sendMessage(sendTo, item.Message.Text)

					if err != nil {
						fmt.Println(err.Error())
						return
					}

					if !result.Ok {
						fmt.Println("Сообщение не отправлено")
					}
				} else {
					result, err := sendMessage(item.Message.Chat.Id, "Пока нет доступных получателей. Оставайтесь на связи!")

					if err != nil {
						fmt.Println(err.Error())
						return
					}

					if !result.Ok {
						fmt.Println("Сообщение не отправлено")
					}
				}

			} else {
				text := processTextResponce(item)

				if text == "" {
					text = "К сожалению, я не понял вашего сообщения. Я еще маленький, только учусь"
				}

				result, err := sendMessage(item.Message.Chat.Id, text)

				if err != nil {
					fmt.Println(err.Error())
					return
				}

				if !result.Ok {
					fmt.Println("Сообщение не отправлено")
				}
			}

		}
	}
}

func processTextResponce(item UpdateResultT) (text string) {

	if item.Message.Text == keywordStart {
		text = "Привет, " + item.Message.From.FirstName + " " + item.Message.From.LastName + "! Рад знакомству! Мне можно сказать привет, спросить как дела, немного поговорить о городах или языках. Особенно люблю говорить об английском. Если уметь хэштег anonym, можно отправить сообщение анонимно. И не забывай говорить пока :) "
	}

	if strings.Contains(strings.ToLower(item.Message.Text), "привет") {
		text = "Привет, " + item.Message.From.FirstName + " " + item.Message.From.LastName
	}

	if strings.Contains(strings.ToLower(item.Message.Text), "дела") {
		text = "Неплохо будет, болтики смазываю. " + item.Message.From.FirstName + ", спасибо что интересуешься "
	}

	if strings.Contains(strings.ToLower(item.Message.Text), "город") {
		if rand.Int()%2 == 0 {
			text = item.Message.From.FirstName + ", хочешь сыграть в города? "
		} else {
			text = "Питер - лучший город, правда комары... болото, сам понимаешь.. "
		}
	}

	if strings.Contains(strings.ToLower(item.Message.Text), "пока") {
		text = "Пока! " + item.Message.From.FirstName + " был рад поболтать. Заходи еще!"
	}

	if strings.Contains(strings.ToLower(item.Message.Text), "язык") || strings.Contains(strings.ToLower(item.Message.Text), "англ") {
		if item.Message.From.Language == "ru" {
			switch rand.Int() % 6 {
			case 0:
				text = item.Message.From.FirstName + " как успехи с Английским?"
			case 1:
				text = "Говорят, фильмы в оригинале смотреть прикольнее. Умеешь?"
			case 2:
				text = "Любишь путешествовать?"
			case 3:
				text = "Читаешь в оригинале?"
			case 4:
				text = "Какая у тебя любимая страна?"
			default:
				text = item.Message.From.FirstName + ", какие языки учишь?"
			}

		} else {
			text = "Ду ю спик фром май хат?"
		}
	}

	return text
}

func getUpdates(LastUpdateId int) (UpdateT, error) {
	url := baseTelegramUrl + "/bot" + telegramToken + "/" + getUpdatesUri + "?offset=" + strconv.Itoa(LastUpdateId+1)
	response := getResponse(url)

	update := UpdateT{}
	err := json.Unmarshal(response, &update)

	if err != nil {
		return update, err
	}
	return update, nil
}

func sendMessage(chatId int, text string) (SendMessageResponseT, error) {
	text = url.QueryEscape(text)
	url := baseTelegramUrl + "/bot" + telegramToken + "/" + sendMessageUri
	url = url + "?chat_id=" + strconv.Itoa(chatId) + "&text=" + text
	response := getResponse(url)

	sendMessage := SendMessageResponseT{}
	err := json.Unmarshal(response, &sendMessage)

	if err != nil {
		return sendMessage, err
	}

	return sendMessage, nil
}

func getResponse(url string) []byte {
	response := make([]byte, 0)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)

		return response
	}

	defer resp.Body.Close()

	for true {
		bs := make([]byte, 1024)
		n, err := resp.Body.Read(bs)
		response = append(response, bs[:n]...)

		if n == 0 || err != nil {
			break
		}
	}

	return response
}
