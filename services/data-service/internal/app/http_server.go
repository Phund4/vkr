package app

import (
	httpHandler "data-service/internal/adapters/http/handlers"
	httpMiddleware "data-service/internal/adapters/http/middleware"
)

// configureHTTPServer применяет обработчик HTTP-ошибок.
func (app *App) configureHTTPServer() {
	app.httpServer.HTTPErrorHandler = httpHandler.HTTPErrorHandler(app.Logger())
	app.httpServer.Use(httpMiddleware.Common()...)
	app.httpServer.Use(httpMiddleware.Logger(app.Logger()))
	app.httpServer.Use(httpMiddleware.ApiV1RateLimiter())
	app.httpServer.Use(httpMiddleware.RoadDataRateLimiter())
}
