package logger

import "github.com/knackwurstking/pgpress/pkg/logger"

// Server

func Server() *logger.Logger     { return logger.GetComponentLogger("Server") }
func Middleware() *logger.Logger { return logger.GetComponentLogger("Middleware") }

// Web Socket Handlers

func WSFeedHandler() *logger.Logger    { return logger.GetComponentLogger("WS Feed Handler") }
func WSFeedConnection() *logger.Logger { return logger.GetComponentLogger("WS Feed Connection") }

// Database Handlers

func DBModifications() *logger.Logger  { return logger.GetComponentLogger("DB Modifications") }
func DBAttachments() *logger.Logger    { return logger.GetComponentLogger("DB Attachments") }
func DBCookies() *logger.Logger        { return logger.GetComponentLogger("DB Cookies") }
func DBFeeds() *logger.Logger          { return logger.GetComponentLogger("DB Feeds") }
func DBTools() *logger.Logger          { return logger.GetComponentLogger("DB Tools") }
func DBUsers() *logger.Logger          { return logger.GetComponentLogger("DB Users") }
func DBMetalSheets() *logger.Logger    { return logger.GetComponentLogger("DB MetalSheets") }
func DBNotes() *logger.Logger          { return logger.GetComponentLogger("DB Notes") }
func DBTroubleReports() *logger.Logger { return logger.GetComponentLogger("DB TroubleReports") }
func DBPressCycles() *logger.Logger    { return logger.GetComponentLogger("DB Service PressCycles") }
func DBRegenerations() *logger.Logger  { return logger.GetComponentLogger("DB Tool Regenerations") }

// HTML Hanlders

func HandlerAuth() *logger.Logger    { return logger.GetComponentLogger("Handler Auth") }
func HandlerFeed() *logger.Logger    { return logger.GetComponentLogger("Handler Feed") }
func HandlerHome() *logger.Logger    { return logger.GetComponentLogger("Handler Home") }
func HandlerProfile() *logger.Logger { return logger.GetComponentLogger("Handler Profile") }
func HandlerTools() *logger.Logger   { return logger.GetComponentLogger("Handler Tools") }

func HandlerTroubleReports() *logger.Logger {
	return logger.GetComponentLogger("Handler TroubleReports")
}

// HTMX Loggers

func HTMXHandlerFeed() *logger.Logger    { return logger.GetComponentLogger("HTMX Handler Feed") }
func HTMXHandlerNav() *logger.Logger     { return logger.GetComponentLogger("HTMX Handler Nav") }
func HTMXHandlerProfile() *logger.Logger { return logger.GetComponentLogger("HTMX Handler Profile") }
func HTMXHandlerTools() *logger.Logger   { return logger.GetComponentLogger("HTMX Handler Tools") }
func HTMXHandlerCycles() *logger.Logger  { return logger.GetComponentLogger("HTMX Handler Cycles") }
func HTMXHandlerTroubleReports() *logger.Logger {
	return logger.GetComponentLogger("HTMX Handler TroubleReports")
}
func HTMXHandlerMetalSheets() *logger.Logger {
	return logger.GetComponentLogger("HTMX Handler MetalSheets")
}
