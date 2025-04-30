package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/go-co-op/gocron/v2"
)

// 消息格式, %d 有且仅有一个
var Format string

var Tomorrow = time.Now()

type MyWriter struct {
	token string
}

func (w MyWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	fmt.Println(strings.Replace(s, w.token, "*********", -1))
	return len(p), nil
}

func main() {
	TOKEN := os.Getenv("TOKEN")
	if TOKEN == "" {
		panic("TOKEN environment variable is empty")
	}

	chatID := os.Getenv("CHAT_ID")
	if chatID == "" {
		panic("CHAT_ID environment variable is empty")
	}

	CHAT_ID, err := strconv.ParseInt(chatID, 10, 64)

	var Last_ID int64 = 0
	var Latest_ID int64 = 0
	var Flag = false

	w := &MyWriter{TOKEN}
	log.SetOutput(w)

	b, err := gotgbot.NewBot(TOKEN, nil)
	if err != nil {
		log.Println("failed to create new bot: " + err.Error())
		return
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	if Format == "" {
		Format = "%d"
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		panic("failed to create new scheduler: " + err.Error())
	}

	t := time.Now()
	Tomorrow = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Add(time.Hour * 24)

	j, err := s.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(0, 0, 0),
			),
		),
		gocron.NewTask(
			func() {
				t = time.Now()
				Tomorrow = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Add(time.Hour * 24)
				if Flag {
					Flag = false
				} else {
					msg := fmt.Sprintf(Format, Latest_ID-Last_ID)
					b.SendMessage(CHAT_ID, msg, nil)
				}
			},
		),
	)

	if err != nil {
		panic("failed to create new job: " + err.Error())
	}

	s.Start()

	dispatcher.AddHandler(handlers.NewCommand("test", func(b *gotgbot.Bot, ctx *ext.Context) error {
		if ctx.EffectiveUser.Id != 1042436080 {
			return nil
		}
		Latest_ID = ctx.EffectiveMessage.MessageId
		j.RunNow()
		return nil
	}))

	dispatcher.AddHandler(handlers.NewCommand("last", func(b *gotgbot.Bot, ctx *ext.Context) error {
		user, err := b.GetChatMember(CHAT_ID, ctx.EffectiveUser.Id, nil)
		if err != nil {
			log.Println(err.Error())
		}

		m := ctx.EffectiveMessage
		if user.GetStatus() != "creator" {
			_, err = m.Reply(b, "非群组创建者", nil)
			if err != nil {
				log.Println(err.Error())
			}
			return nil
		}

		c := strings.SplitN(m.Text, " ", 3)
		var id int64 = 0
		if len(c) != 2 {
			if m.ReplyToMessage != nil {
				id = m.ReplyToMessage.MessageId
			} else {
				id = m.MessageId
			}
		} else {
			id, err = strconv.ParseInt(c[1], 10, 64)

			if err != nil {
				_, err = m.Reply(b, "无效参数", nil)
				if err != nil {
					log.Println(err.Error())
				}
				return nil
			}
			if id > m.MessageId {
				msg := fmt.Sprintf("%d 大于当前最新的消息ID", id)
				_, err = m.Reply(b, msg, nil)
				if err != nil {
					log.Println(err.Error())
				}
				return nil
			}

			if id < 0 {
				msg := fmt.Sprintf("%d 不可为负数", id)
				_, err = m.Reply(b, msg, nil)
				if err != nil {
					log.Println(err.Error())
				}
				return nil
			}
		}
		Last_ID = id
		msg := fmt.Sprintf("Last_ID 已设置为 %d", Last_ID)
		_, err = m.Reply(b, msg, nil)
		if err != nil {
			log.Println(err.Error())
		}
		return nil
	}))

	dispatcher.AddHandler(handlers.NewMessage(message.All, func(b *gotgbot.Bot, ctx *ext.Context) error {
		if ctx.EffectiveChat.Id != CHAT_ID {
			return nil
		}
		msg := ctx.EffectiveMessage
		Latest_ID = msg.MessageId
		if Last_ID == 0 {
			Last_ID = Latest_ID - 1
		}

		log.Printf("%d %d\n", Latest_ID, Last_ID)

		// ~~下面那段虽然估计没必要，但万一呢~~
		t = time.Unix(msg.GetDate(), 0)
		if !t.Before(Tomorrow) {
			Flag = true
			msg := fmt.Sprintf(Format, Latest_ID-Last_ID)
			b.SendMessage(CHAT_ID, msg, nil)
			Last_ID = Latest_ID
			Tomorrow = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Add(time.Hour * 24)
		}
		return nil
	}))

	err = updater.StartPolling(b, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", b.User.Username)

	updater.Idle()
}
