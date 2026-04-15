package utils

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vinifrancosilva/lista_v2/internal/models"
)

func VerificaSessao(c echo.Context) (models.Usuario, error) {
	sessao, err := session.Get("sessao", c)
	if err != nil {
		return models.Usuario{}, err
	}

	// testa se existe usuario na sessao atual, ou seja, se está logado e se não é o path /login
	_, ok := sessao.Values["usuario_id"]

	if !ok {
		return models.Usuario{}, nil
	}

	// cria usuario struct com o ID
	usuario := models.Usuario{ID: sessao.Values["usuario_id"].(int32)}

	return usuario, nil
}
