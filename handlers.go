package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vinifrancosilva/lista_v2/models"
	"github.com/vinifrancosilva/lista_v2/web/components/login"
	"github.com/vinifrancosilva/lista_v2/web/pages"

	datastar "github.com/starfederation/datastar/sdk/go"
)

//[ ] TODO: ao compartilhar a lista com outro usuario, compartilha também as categorias

// Custom Middlewares
// Testa de esta logado
func middlewareEstaLogado(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sessao, err := session.Get("sessao", c)
		if err != nil {
			return err
		}

		// testa se existe usuario na sessao atual, ou seja, se está logado e se não é o path /login
		_, ok := sessao.Values["usuario"]
		if !ok && c.Request().URL.Path != "/login" {
			// redireciona para a pagina de login caso não esteja logado
			return c.Redirect(http.StatusFound, "/login")
		}

		// se já está logado e está tentando entrar na página de login, rediciona pra index
		if ok && c.Request().URL.Path == "/login" {
			return c.Redirect(http.StatusFound, "/")
		}

		return next(c)
	}
}

// Handlers dos endpoints
func handlerIndex(c echo.Context) error {
	// TODO: Implement handlerIndex
	// pages.Index("Northstar").Render(r.Context(), w)
	return Render(c, http.StatusOK, pages.Index("Listas V2"))
}

func handlerLoginPage(c echo.Context) error {
	// TODO: Implement handlerLoginPage
	return Render(c, http.StatusOK, pages.LoginPage())
}

func handlerLoginPost(c echo.Context) error {
	// Instancia o model sqlc
	m := models.New(dbPool)

	// nesse caso não precisei criar uma struct pro signal, a struct do sqlc já serve
	paramsFromSignals := &models.TestaLoginParams{}

	// faz o marshall usando o bind do echo
	// err := c.Bind(signals)
	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// faz o marshall usando o sdk datastar
	if err := datastar.ReadSignals(c.Request(), paramsFromSignals); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Cria os parametros para a query com os valores recebidos via post

	// Testa o login no banco
	usuario, err := m.TestaLogin(context.Background(), *paramsFromSignals)
	// se deu erro no login...
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("autenticação falhou: %w", err)
	}
	if err == pgx.ErrNoRows {
		datastar.NewSSE(c.Response().Writer, c.Request()).MergeFragmentTempl(login.MsgErro("Usuário ou senha inválidos"))

		return c.NoContent(http.StatusOK)
	}

	// Login com sucesso
	// cria a sessão
	sessao, err := session.Get("sessao", c)
	if err != nil {
		return err
	}

	// adiciona o usuario
	sessao.Values["usuario_id"] = usuario.ID
	sessao.Values["usuario"] = usuario.Usuario

	// salva a sessão
	if err = sessao.Save(c.Request(), c.Response().Writer); err != nil {
		return err
	}

	// redireciona para a index
	return datastar.NewSSE(c.Response().Writer, c.Request()).Redirect("/")
}

func handlerLogout(c echo.Context) error {
	// chama a sessão
	sessao, err := session.Get("sessao", c)
	if err != nil {
		return err
	}

	// deleta a session.
	sessao.Options.MaxAge = -1
	if err = sessao.Save(c.Request(), c.Response().Writer); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/login")
	// return datastar.NewSSE(c.Response().Writer, c.Request()).Redirect("/login")
}

// Função auxiliar para renderizar templates do templ com o echo
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
