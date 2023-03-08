package settings

import (
	"csc-code-test/internal/logger"
	"go.uber.org/zap"
	"os"
	"strconv"
)

var Settings struct {
	SecondsToWait int
}

func LoadSettings() {
	envSeconds := os.Getenv("ENV_SECONDS")

	secondsToWait, err := strconv.ParseInt(envSeconds, 10, 32)
	if err != nil || secondsToWait == 0 {
		logger.Logger.Warn("invalid ENV_SECONDS value", zap.String("envSeconds", envSeconds))
		secondsToWait = 10
	}

	Settings.SecondsToWait = int(secondsToWait)

	logger.Logger.Info("SecondsToWait", zap.Int("SecondsToWait", Settings.SecondsToWait))

}
