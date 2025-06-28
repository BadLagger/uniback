package main

import (
	"uniback/utils"
)

func main() {
	logger := utils.GlobalLogger().SetLevel(utils.Debug)

	logger.Info("Start APP!!!")
	defer logger.Info("APP Done!!!")

	cfg := utils.CfgLoad("UniBack")
	logger.Info("Set app name: %s", cfg.AppName)
}
