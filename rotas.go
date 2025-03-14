package main

import (
	"github.com/labstack/echo/v4"
)

func defineRotas(e *echo.Echo) {
	// Index
	e.GET("/", handlerIndex)

	// Login / Logout
	e.GET("/login", handlerLoginPage)
	e.POST("/login", handlerLoginPost)
	e.GET("/logout", handlerLogout)

	// API Group
	g := e.Group("/api")

	// API da Lista
	g.GET("/lista", handlerApiListaGet)
	g.POST("/lista", handlerFake)
	g.PATCH("/lista", handlerFake)
	g.DELETE("/lista", handlerFake)

}
