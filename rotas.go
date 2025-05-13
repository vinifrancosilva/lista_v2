package main

import (
	"github.com/labstack/echo/v4"
)

func defineRotas(e *echo.Echo) {
	// Static
	e.Static("/static", "static")

	// Index
	e.GET("/", handlerIndex)

	// Login / Logout
	e.GET("/login", handlerLoginPage)
	e.POST("/login", handlerLoginPost)
	e.GET("/logout", handlerLogout)

	// Categorias
	e.GET("/categorias", handlerFake)

	// API Group
	g := e.Group("/api")

	// API da Lista
	g.GET("/lista", handlerApiListaGet)
	g.POST("/lista", handlerApiListaPost)
	g.PATCH("/lista/:id", handlerApiListaPatch)
	g.DELETE("/lista/:id", handlerApiListaDelete)

}
