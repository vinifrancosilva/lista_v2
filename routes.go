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
}
