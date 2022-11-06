package main

import (
	"context"
	"github.com/reijo1337/ToxicBot/internal/google_spreadsheet"
	"github.com/reijo1337/ToxicBot/internal/storage"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_sticker"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_text"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_text/bulling"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_text/igor"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_user_join"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_user_left"
	"github.com/reijo1337/ToxicBot/internal/handlers/on_voice"
	"github.com/reijo1337/ToxicBot/internal/utils"
	"gopkg.in/telebot.v3"

	"github.com/sirupsen/logrus"
)

type config struct {
	TelegramToken           string        `envconfig:"TELEGRAM_TOKEN" required:"true"`
	TelegramLongPollTimeout time.Duration `envconfig:"TELEGRAM_LONG_POLL_TIMEOUT" default:"10s"`
	StickerSets             []string      `envconfig:"STICKER_SETS" default:"static_bulling_by_stickersthiefbot"`
}

func main() {
	ctx := context.Background()

	logger := newLogger()

	cfg, err := newConfig()
	if err != nil {
		logger.WithError(err).Fatal("can't init config")
	}

	gs, err := google_spreadsheet.New(ctx)
	if err != nil {
		logger.WithError(err).Fatal("can't create google spreadsheet instance")
	}

	stor := storage.New(gs)

	pref := telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: cfg.TelegramLongPollTimeout},
		OnError: func(err error, ctx telebot.Context) {
			logger.
				WithError(err).
				WithField("update", ctx.Update()).
				Error("can't handle update")
		},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		logger.WithError(err).Fatal("can't init bot api")
	}

	igorHandler, err := igor.New(stor)
	if err != nil {
		logger.WithError(err).Fatal("init on_text igor handler")
	}

	bullingHandler, err := bulling.New(ctx, stor, logger)
	if err != nil {
		logger.WithError(err).Fatal("init on_text bulling handler")
	}

	b.Handle(
		telebot.OnText,
		on_text.New(
			igorHandler,
			bullingHandler,
		).Handle,
	)

	greetingsHandler, err := on_user_join.New(ctx, stor, logger)
	if err != nil {
		logger.WithError(err).Fatal("can't init on_user_join handler")
	}
	b.Handle(telebot.OnUserJoined, greetingsHandler.Handle)

	b.Handle(telebot.OnUserLeft, on_user_left.Handle)

	stickersFromPacks := []string{}
	if len(cfg.StickerSets) > 0 {
		stickersFromPacks, err = utils.GetStickersFromPacks(b, cfg.StickerSets)
		if err != nil {
			logger.WithError(err).Warn("can't get stickers from sticker packs")
		}
	}

	stickersReactionHandler, err := on_sticker.New(ctx, stor, logger, stickersFromPacks)
	if err != nil {
		logger.WithError(err).Fatal("can't init on_sticker handler")
	}

	b.Handle(telebot.OnSticker, stickersReactionHandler.Handle)

	onVoice, err := on_voice.New(ctx, stor, logger)
	if err != nil {
		logger.WithError(err).Fatal("can't init on_voice handler")
	}
	b.Handle(telebot.OnVoice, onVoice.Handle)

	go func() {
		logger.WithField("user_name", b.Me.Username).Info("bot started")
		b.Start()
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	b.Stop()
}

func newLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{})
	logger.SetReportCaller(true)

	return logger
}

func newConfig() (*config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		envconfig.Usage("", cfg)
		return nil, err
	}

	return &cfg, nil
}
