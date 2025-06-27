package main

import (
	"uniback/utils"
)

func main() {
	logger := utils.NewLogger().SetLevel(utils.All)

	logger.Info("Start APP!!!")
	defer logger.Info("App Done!!!")

	count := 1
	logger.Debug("Debug message %d", count)
	count += 1

	logger.Debug("Debug message %d", count)

	count += 1
	logger.Critical("Debug message %d", count)
}
