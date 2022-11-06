package bulling

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	ThresholdCount       int           `envconfig:"BULLINGS_THRESHOLD_COUNT" default:"5"`
	ThresholdTime        time.Duration `envconfig:"BULLINGS_THRESHOLD_TIME" default:"1m"`
	Cooldown             time.Duration `envconfig:"BULLINGS_COOLDOWN" default:"1h"`
	MarkovChance         float32       `envconfig:"BULLINGS_MARKOV_CHANCE" default:"0.75"`
	UpdateMessagesPeriod time.Duration `envconfig:"BULLINGS_UPDATE_MESSAGES_PERIOD" default:"10m"`
}

func (b *bulling) parseConfig() error {
	if err := envconfig.Process("", &b.cfg); err != nil {
		envconfig.Usage("", b.cfg)
		return err
	}

	return nil
}
