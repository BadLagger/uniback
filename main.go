package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"uniback/repository/postgres"
	"uniback/utils"
)

func main() {
	logger := utils.GlobalLogger().SetLevel(utils.Debug)

	logger.Info("Start APP!!!")
	defer logger.Info("APP Done!!!")

	cfg := utils.CfgLoad("UniBack")
	logger.Info("Set app name: %s", cfg.AppName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-quit:
			logger.Info("Signal to escape! Shutdown")
		case <-ctx.Done():
			logger.Error("Context DONE! It wouldn't be happened!!!")
		}
	}()

	DataBase := postgres.New(ctx, postgres.PgConfigFromConfig(*cfg))

	if DataBase == nil {
		logger.Critical("Database error")
		return
	}
	defer DataBase.Close()
}
