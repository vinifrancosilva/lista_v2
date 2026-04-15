package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vinifrancosilva/lista_v2/internal/utils"
)

// Testa de esta logado
func MiddlewareEstaLogado(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// se for /static não faz a verificação de login
		if len(c.Request().URL.Path) >= 7 && c.Request().URL.Path[0:7] == "/static" {
			return next(c)
		}
		// verifica se existe sessão
		usuario, err := utils.VerificaSessao(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// se não existe sessão, não está logado, redireciona pra login
		if usuario.ID == 0 && c.Request().URL.Path != "/login" {
			// redireciona para a pagina de login caso não esteja logado
			return c.Redirect(http.StatusFound, "/login")
		}

		// se já está logado e está tentando entrar na página de login, rediciona pra index
		if usuario.ID > 0 && c.Request().URL.Path == "/login" {
			return c.Redirect(http.StatusFound, "/")
		}

		return next(c)
	}
}
