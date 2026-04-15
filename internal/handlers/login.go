package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/vinifrancosilva/lista_v2/internal/models"
	"github.com/vinifrancosilva/lista_v2/views/components/login"
	"github.com/vinifrancosilva/lista_v2/views/pages"
)

func HandlerLoginPage(c echo.Context) error {
	return Render(c, http.StatusOK, pages.LoginPage())
}

func HandlerLoginPost(c echo.Context) error {
	// nesse caso não precisei criar uma struct pro signal, a struct do sqlc já serve
	var paramsFromSignals struct {
		Usuario string `json:"usuario"`
		Senha   string `json:"senha"`
	}

	// faz o marshall usando o sdk datastar
	if err := datastar.ReadSignals(c.Request(), &paramsFromSignals); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Cria os parametros para a query com os valores recebidos via post
	usuario := models.Usuario{
		Usuario: paramsFromSignals.Usuario,
		Senha:   paramsFromSignals.Senha,
	}

	// testa o login no banco
	err := usuario.TestaLogin(context.Background())
	if err != nil && err != pgx.ErrNoRows {
		// se o erro for genérico, deu erro na comunicação com o banco
		sse := datastar.NewSSE(c.Response().Writer, c.Request())
		sse.PatchElementTempl(
			login.MsgErro(
				fmt.Sprintf("autenticação falhou: %v", err),
			),
			datastar.WithUseViewTransitions(true),
		)

		return nil
		//return c.NoContent(http.StatusOK)
	}
	if err == pgx.ErrNoRows {
		// se o erro for NoRows, foi passado usuário e senha inválido
		sse := datastar.NewSSE(c.Response().Writer, c.Request())
		sse.PatchElementTempl(
			login.MsgErro("Usuário ou senha inválidos"),
			datastar.WithUseViewTransitions(true),
		)

		return nil
		//return c.NoContent(http.StatusOK)
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
	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	return sse.Redirect("/")

}

func HandlerLogout(c echo.Context) error {
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
}
