package htmx

import (
	"net/http"
	"time"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/logger"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/wshandlers"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

type Nav struct {
	DB            *database.DB
	WSFeedHandler *wshandlers.FeedHandler
}

func (h *Nav) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/nav/feed-counter", h.handleFeedCounterWebSocketEcho),
		},
	)
}

// handleFeedCounterWebSocketEcho creates an echo-compatible WebSocket handler
func (h *Nav) handleFeedCounterWebSocketEcho(c echo.Context) error {
	remoteIP := c.RealIP()
	userAgent := c.Request().UserAgent()
	origin := c.Request().Header.Get("Origin")

	start := time.Now()
	logger.HTMXHandlerNav().Info("WebSocket upgrade request from %s (user-agent: %s, origin: %s)",
		remoteIP, userAgent, origin)

	// Create a WebSocket handler that can work with Echo
	wsHandler := websocket.Handler(func(ws *websocket.Conn) {
		connectionStart := time.Now()

		// Get user from echo context
		user, err := webhelpers.GetUserFromContext(c)
		if err != nil {
			logger.HTMXHandlerNav().Error("WebSocket authentication failed from %s: %v", remoteIP, err)
			ws.Close()
			return
		}

		logger.HTMXHandlerNav().Info("WebSocket user authenticated: %s (ID: %d, LastFeed: %d) from %s",
			user.Name, user.TelegramID, user.LastFeed, remoteIP)

		// Log connection details
		localAddr := ws.LocalAddr()
		logger.HTMXHandlerNav().Debug("WebSocket connection established: %s -> %s for user %s",
			remoteIP, localAddr, user.Name)

		// Register the connection with the feed notifier
		registrationStart := time.Now()
		feedConn := h.WSFeedHandler.RegisterConnection(user.TelegramID, user.LastFeed, ws)
		registrationElapsed := time.Since(registrationStart)

		if feedConn == nil {
			logger.HTMXHandlerNav().Error("Failed to register WebSocket connection for user %s from %s (registration took: %v)",
				user.Name, remoteIP, registrationElapsed)
			ws.Close()
			return
		}

		connectionElapsed := time.Since(connectionStart)
		logger.HTMXHandlerNav().Info("WebSocket connection registered for user %s from %s in %v (registration: %v)",
			user.Name, remoteIP, connectionElapsed, registrationElapsed)

		// Track active connections
		defer func() {
			totalDuration := time.Since(connectionStart)
			logger.HTMXHandlerNav().Info("WebSocket connection closed for user %s from %s after %v",
				user.Name, remoteIP, totalDuration)
		}()

		// Start the write pump in a goroutine
		logger.HTMXHandlerNav().Debug("Starting write pump for user %s from %s", user.Name, remoteIP)
		go feedConn.WritePump()

		// Start the read pump (this will block until connection closes)
		logger.HTMXHandlerNav().Debug("Starting read pump for user %s from %s", user.Name, remoteIP)
		feedConn.ReadPump(h.WSFeedHandler)

		logger.HTMXHandlerNav().Debug("Read pump finished for user %s from %s", user.Name, remoteIP)
	})

	// Measure WebSocket setup time
	setupElapsed := time.Since(start)
	logger.HTMXHandlerNav().Debug("WebSocket handler setup completed in %v for %s", setupElapsed, remoteIP)

	// Serve the WebSocket connection
	wsHandler.ServeHTTP(c.Response(), c.Request())

	totalElapsed := time.Since(start)
	logger.HTMXHandlerNav().Debug("WebSocket request handling completed in %v for %s", totalElapsed, remoteIP)

	return nil
}
