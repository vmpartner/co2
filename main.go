package main

import (
	"bufio"
	"crypto/tls"
	"github.com/astaxie/beego/logs"
	"github.com/jinzhu/gorm"
	"github.com/tarm/serial"
	"github.com/tv42/topic"
	"github.com/vmpartner/go-tools"
	"gopkg.in/ini.v1"
	"gopkg.in/tucnak/telebot.v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type App struct {
	DB         *gorm.DB
	Bot        *telebot.Bot
	Config     *ini.File
	Topic      *topic.Topic
	UserCh     map[int]chan interface{}
	UserActive map[int]bool
	UserSleep  map[int]time.Duration
}

func main() {

	// Init
	app := new(App)
	app.Topic = topic.New()
	app.UserCh = make(map[int]chan interface{})
	app.UserActive = make(map[int]bool)
	app.UserSleep = make(map[int]time.Duration)

	// Config
	var err error
	app.Config, err = ini.Load("conf/app.conf")
	tools.CheckErr(err)

	// Init bot connection
	app.Bot, err = telebot.NewBot(telebot.Settings{
		Token:  app.Config.Section("telegram").Key("token").String(),
		URL:    app.Config.Section("telegram").Key("url").String(),
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		Client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}},
	})
	tools.CheckErr(err)

	// Read serial port
	go func(app *App) {
		port := app.Config.Section("serial").Key("port").String()
		baud, _ := app.Config.Section("serial").Key("baud").Int()
		timeout, _ := app.Config.Section("serial").Key("timeout").Int()
		c := &serial.Config{Name: port, Baud: baud, ReadTimeout: time.Second * time.Duration(timeout)}
		s, _ := serial.OpenPort(c)
		defer func() {
			err := s.Close()
			tools.CheckErr(err)
		}()
		k := 0.0
		for {
			k += 0.001
			if k <= 0.0010 {
				continue
			}
			scanner := bufio.NewScanner(s)
			for scanner.Scan() {
				value := scanner.Text()
				logs.Info("CO2:", value)
				app.Topic.Broadcast <- value
				break
			}
		}
	}(app)

	// Telegram bot
	go func(app *App) {

		userID, _ := app.Config.Section("telegram").Key("user").Int()
		if userID > 0 {
			go app.SubscribeToAlert(userID)
		}

		// Start watch alert
		app.Bot.Handle("/start", func(m *telebot.Message) {
			go app.SubscribeToAlert(m.Sender.ID)
		})

		// Stop watch alert
		app.Bot.Handle("/stop", func(m *telebot.Message) {
			app.Topic.Unregister(app.UserCh[m.Sender.ID])
			app.UserActive[m.Sender.ID] = false
			app.UserSleep[m.Sender.ID] = time.Duration(0)
			close(app.UserCh[m.Sender.ID])
		})

		// Sleep
		app.Bot.Handle("/sleep", func(m *telebot.Message) {
			s := strings.Replace(m.Text, "/sleep", "", -1)
			s = strings.TrimSpace(s)
			i, _ := strconv.Atoi(s)
			sleep := time.Duration(i) * time.Minute
			logs.Info("SLEEP ", sleep)
			app.UserSleep[m.Sender.ID] = sleep
		})

		app.Bot.Start()
	}(app)

	// Loop
	select {}
}

func (app *App) SubscribeToAlert(userID int) {
	app.UserActive[userID] = true
	for {
		if app.UserActive[userID] == false {
			break
		}
		app.UserCh[userID] = make(chan interface{})
		app.Topic.Register(app.UserCh[userID])
		for value := range app.UserCh[userID] {
			if app.UserSleep[userID] > 0 {
				time.Sleep(app.UserSleep[userID])
				app.UserSleep[userID] = time.Duration(0)
				break
			}
			i, err := strconv.Atoi(value.(string))
			valueStr := value.(string)
			tools.CheckErr(err)
			warnValue, _ := app.Config.Section("values").Key("warn").Int()
			goodValue, _ := app.Config.Section("values").Key("good").Int()
			if i >= warnValue {
				logs.Warn("SEND ALERT: ", value)
				user := telebot.User{}
				user.ID = userID
				_, err := app.Bot.Send(&user, "🔥 CO2 BAD "+valueStr)
				tools.CheckErr(err)
				timeout, _ := app.Config.Section("values").Key("timeout").Int()
				sleep := time.Duration(timeout) * time.Minute
				logs.Info("SLEEP ", sleep)
				time.Sleep(sleep)
			}
			if i <= goodValue {
				logs.Warn("SEND OK: ", value)
				user := telebot.User{}
				user.ID = userID
				_, err := app.Bot.Send(&user, "🌍 CO2 GOOD "+valueStr)
				tools.CheckErr(err)
				timeout, _ := app.Config.Section("values").Key("timeout").Int()
				sleep := time.Duration(timeout*2) * time.Minute
				logs.Info("SLEEP ", sleep)
				time.Sleep(sleep)
			}
		}
	}
}
