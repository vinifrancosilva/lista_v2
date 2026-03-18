package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/vinifrancosilva/lista_v2/internal/models"
	"github.com/vinifrancosilva/lista_v2/internal/utils"
	lc "github.com/vinifrancosilva/lista_v2/views/components/listas"
	"github.com/vinifrancosilva/lista_v2/views/components/login"
	"github.com/vinifrancosilva/lista_v2/views/pages"

	datastar "github.com/starfederation/datastar/sdk/go"
)

//[ ] TODO: ao compartilhar a lista com outro usuario, compartilha também as categorias

// Custom Middlewares
// Testa de esta logado
func middlewareEstaLogado(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// se for /static não faz a verificação de login
		if len(c.Request().URL.Path) >= 7 && c.Request().URL.Path[0:7] == "/static" {
			return next(c)
		}
		// verifica se existe sessão
		usuario_id, err := utils.VerificaSessao(c)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// se não existe sessão, não está logado, redireciona pra login
		if usuario_id == 0 && c.Request().URL.Path != "/login" {
			// redireciona para a pagina de login caso não esteja logado
			return c.Redirect(http.StatusFound, "/login")
		}

		// se já está logado e está tentando entrar na página de login, rediciona pra index
		if usuario_id > 0 && c.Request().URL.Path == "/login" {
			return c.Redirect(http.StatusFound, "/")
		}

		return next(c)
	}
}

// Handlers dos endpoints
func handlerIndex(c echo.Context) error {
	// pega usuario da sessao
	usuario_id, err := utils.VerificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// cria struct do usuário da sessão
	usuario := models.Usuario{
		ID: usuario_id,
	}

	// pega possíveis usuários para compartilhar
	usuariosCompartilhamento, err := usuario.PegaUsuariosParaCompartilhar(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return Render(c, http.StatusOK, pages.Index("Listas V2", usuariosCompartilhamento))
}

func handlerLoginPage(c echo.Context) error {
	return Render(c, http.StatusOK, pages.LoginPage())
}

func handlerLoginPost(c echo.Context) error {
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
		datastar.NewSSE(
			c.Response().Writer,
			c.Request(),
		).MergeFragmentTempl(
			login.MsgErro(
				fmt.Sprintf("autenticação falhou: %v", err),
			),
			datastar.WithUseViewTransitions(true),
		)

		return c.NoContent(http.StatusOK)
	}
	if err == pgx.ErrNoRows {
		// se o erro for NoRows, foi passado usuário e senha inválido
		datastar.NewSSE(
			c.Response().Writer,
			c.Request(),
		).MergeFragmentTempl(
			login.MsgErro("Usuário ou senha inválidos"),
			datastar.WithUseViewTransitions(true),
		)

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

// Handler fake só pra deixar as rotas montadas
func handlerFake(c echo.Context) error {
	return c.String(http.StatusOK, "FAKE HANDLER")
}

// ----------------------- API -----------------------

// Lista
// Get - abre o canal SSE que recebe todas as mudancas realizadas na lista
func handlerApiListaGet(c echo.Context) error {
	// pega usuario da sessao
	usuario_id, err := utils.VerificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// cria a struct desse endpoint
	ch := make(chan struct{})
	subs := models.Subscriber{
		Endpoint: "listas",
		Channel:  ch,
	}

	// abre o canal SSE
	sse := datastar.NewSSE(c.Response().Writer, c.Request())

	// faz o subscribe
	subscriberChan <- subs

	for {
		select {
		// se fechar a conexão, faz o unsubscribe
		case <-c.Request().Context().Done():
			// fecha o canal
			unsubscriberChan <- subs
			return nil
		case <-ch:
			// antes de enviar qualquer update, verifica se a sessão ainda está válida
			usuario_id_check, err := utils.VerificaSessao(c)
			if err != nil {
				unsubscriberChan <- subs
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			// se não estiver válida, redireciona pra login
			if usuario_id_check == 0 {
				unsubscriberChan <- subs
				sse.Redirect("/login")
				return nil
			}

			// se estiver válida, pega as listas atualizadas
			usuario := models.Usuario{ID: usuario_id}
			listas, err := usuario.PegaListas(c.Request().Context())
			if err != nil {
				unsubscriberChan <- subs
				sse.Redirect("/login")
				return nil
			}

			// envia pro front com sse.MergeFragmentTempl()
			vt := datastar.WithUseViewTransitions(true)
			sse.MergeFragmentTempl(lc.ListaDeListas(listas), vt)
		}
	}
}

func handlerApiListaPost(c echo.Context) error {
	// pega usuario da sessao
	usuario_id, err := verificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// para manter separation os concerns, cria uma DTO pra receber e validar info do front
	listaPostDTO := struct {
		Lista     string `json:"lista" validate:"required"`
		Descricao string `json:"descricao"`
	}{}

	// faz o parse dos dados do datastar
	if err := datastar.ReadSignals(c.Request(), &listaPostDTO); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// valida o DTO
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(&listaPostDTO)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			sse := datastar.NewSSE(c.Response().Writer, c.Request())
			return sse.MergeSignals(
				[]byte(`{input_lista_erro: "Nome da lista é necessário..."}`),
			)
		}
	}

	// cria usuario struct com o ID
	usuario := models.Usuario{ID: usuario_id}

	// insere no banco as informações validadas
	err = usuario.InsereLista(c.Request().Context(), listaPostDTO.Lista, listaPostDTO.Descricao)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// limpa msg de erro caso exista
	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	limpa_erro := `{input_lista_erro: ''}`
	sse.MergeSignals([]byte(limpa_erro))

	// manda evento pra publicação
	publisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

func handlerApiListaDelete(c echo.Context) error {
	// pega usuario da sessao
	usuario_id, err := verificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// pega o ID da lista do path
	listaID := c.Param("id")
	var listaIDInt32 int32
	_, err = fmt.Sscanf(listaID, "%d", &listaIDInt32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID inválido")
	}

	// faz o delete
	err = models.DeletaLista(c.Request().Context(), listaIDInt32, usuario_id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// manda evento pra publicação
	publisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

func handlerApiListaPatch(c echo.Context) error {
	SignalsAndParams := struct {
		ListaID   int32  `param:"id"`
		Lista     string `json:"lista"`
		Descricao string `json:"descricao"`
	}{}

	// pega usuario da sessao
	usuario_id, err := verificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// pega parâmetros do path
	err = c.Bind(&SignalsAndParams)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// faz o update
	err = models.AtualizaLista(c.Request().Context(), SignalsAndParams.ListaID, usuario_id, SignalsAndParams.Lista, SignalsAndParams.Descricao)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// manda evento pra publicação
	publisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

// TODO: migrar o daisyui/tailwind de CDN para o node pra conseguir usar o folha de modelo para o choicesjs
