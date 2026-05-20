package app

import (
	httpHandler "pusher/internal/adapters/http/handlers"
	httpMiddleware "pusher/internal/adapters/http/middleware"
)

func (app *App) configureHTTPServer() {
	app.httpServer.HTTPErrorHandler = httpHandler.HTTPErrorHandler(app.Logger())
	app.httpServer.Use(httpMiddleware.Common()...)
	app.httpServer.Use(httpMiddleware.Logger(app.Logger()))
}
