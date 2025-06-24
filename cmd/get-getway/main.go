package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/Koyo-os/get-getway/internal/config"
	"github.com/Koyo-os/get-getway/internal/core"
	"github.com/Koyo-os/get-getway/internal/entity"
	"github.com/Koyo-os/get-getway/internal/transport/consumer"
	"github.com/Koyo-os/get-getway/internal/transport/getserver"
	"github.com/Koyo-os/get-getway/pkg/listener"
	"github.com/Koyo-os/get-getway/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func getLogger() *logger.Logger {
	loggerCFG := logger.Config{
		LogFile:   "app.log",
		LogLevel:  "debug",
		AppName:   "get-gateway",
		AddCaller: true,
	}

	if err := logger.Init(loggerCFG); err != nil {
		fmt.Print(err)

		return nil
	}

	return logger.Get()
}

func getCore(urls map[string]string) (*core.GetServiceCore, error) {
	conn, err := core.ConnectToServices(urls)
	if err != nil {
		return nil, err
	}

	core := core.NewGetServiceCore(conn)
	return core, nil
}

func runGetServer(core *core.GetServiceCore, addr string) error {
	getserver := getserver.NewGetServer(core)

	return getserver.Run(context.Background(), addr)
}

func runConsumer(cfg *config.Config, output chan entity.Event, ctx context.Context) error {
	conn, err := amqp.Dial(cfg.Urls["rabbitmq"])
	if err != nil{
		return err
	}

	consumer, err := consumer.NewConsumer(cfg, logger.Get(), conn)
	if err != nil{
		return err
	}

	go consumer.ConsumeMessages(output, ctx)

	return nil
}

func runListener(input chan entity.Event, core *core.GetServiceCore, ctx context.Context) {
	listener := listener.NewListener(input, core)

	go func(){
		listener.Listen(ctx)
	}()
}

func main() {
	eventChan := make(chan entity.Event, 5) 

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := config.NewConfig()
	logger := getLogger()

	core, err := getCore(cfg.Urls)
	if err != nil {
		logger.Error("failed get core", zap.Error(err))

		return
	}

	if err := runGetServer(core, fmt.Sprintf("localhost:%d", cfg.Port)); err != nil {
		logger.Error("failed run get-server", zap.Error(err))

		return
	}

	if err := runConsumer(cfg, eventChan, ctx);err != nil{
		logger.Error("failed run consumer", zap.Error(err))

		return
	}

	runListener(eventChan, core, ctx)

	logger.Info("all get-gateway is healthy and ready to work!")
}
