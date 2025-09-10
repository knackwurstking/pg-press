package logger

import "github.com/knackwurstking/pgpress/pkg/logger"

func Middleware() *logger.Logger          { return logger.GetComponentLogger("Middleware") }
func Server() *logger.Logger              { return logger.GetComponentLogger("Server") }
func WSFeedHandler() *logger.Logger       { return logger.GetComponentLogger("WS Feed Handler") }
func WSFeedConnection() *logger.Logger    { return logger.GetComponentLogger("WS Feed Connection") }
func DBAttachments() *logger.Logger       { return logger.GetComponentLogger("DB Attachments") }
func DBCookies() *logger.Logger           { return logger.GetComponentLogger("DB Cookies") }
func DBFeeds() *logger.Logger             { return logger.GetComponentLogger("DB Feeds") }
func DBTools() *logger.Logger             { return logger.GetComponentLogger("DB Tools") }
func DBUsers() *logger.Logger             { return logger.GetComponentLogger("DB Users") }
func DBMetalSheets() *logger.Logger       { return logger.GetComponentLogger("DB MetalSheets") }
func DBNotes() *logger.Logger             { return logger.GetComponentLogger("DB Notes") }
func DBTroubleReports() *logger.Logger    { return logger.GetComponentLogger("DB TroubleReports") }
func DBPressCycles() *logger.Logger       { return logger.GetComponentLogger("DB Service PressCycles") }
func DBToolRegenerations() *logger.Logger { return logger.GetComponentLogger("DB ToolRegenerations") }
func HandlerAuth() *logger.Logger         { return logger.GetComponentLogger("Handler Auth") }
func HandlerFeed() *logger.Logger         { return logger.GetComponentLogger("Handler Feed") }
func HandlerHome() *logger.Logger         { return logger.GetComponentLogger("Handler Home") }
func HandlerProfile() *logger.Logger      { return logger.GetComponentLogger("Handler Profile") }
func HandlerTools() *logger.Logger        { return logger.GetComponentLogger("Handler Tools") }

func HandlerTroubleReports() *logger.Logger {
	return logger.GetComponentLogger("Handler TroubleReports")
}

func HTMXHandlerFeed() *logger.Logger    { return logger.GetComponentLogger("HTMX Handler Feed") }
func HTMXHandlerNav() *logger.Logger     { return logger.GetComponentLogger("HTMX Handler Nav") }
func HTMXHandlerProfile() *logger.Logger { return logger.GetComponentLogger("HTMX Handler Profile") }
func HTMXHandlerTools() *logger.Logger   { return logger.GetComponentLogger("HTMX Handler Tools") }

func HTMXHandlerTroubleReports() *logger.Logger {
	return logger.GetComponentLogger("HTMX Handler TroubleReports")
}
