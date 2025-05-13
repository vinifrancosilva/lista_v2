package main

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func verificaSessao(c echo.Context) (int32, error) {
	sessao, err := session.Get("sessao", c)
	if err != nil {
		return 0, err
	}

	// testa se existe usuario na sessao atual, ou seja, se está logado e se não é o path /login
	_, ok := sessao.Values["usuario_id"]

	if !ok {
		return 0, nil
	}

	return sessao.Values["usuario_id"].(int32), nil
}
