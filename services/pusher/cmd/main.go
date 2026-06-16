package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	application "pusher/internal/app"
)

func main() {
	ctx, cancel := ctxWithSyscallHandler()
	defer cancel()

	app := application.New(ctx)
	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().Err(err).Msg("pusher stopped with error")
	}
}

func ctxWithSyscallHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}
