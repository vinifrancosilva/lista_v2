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
		// se for /static não faz a verificação de login
		if len(c.Request().URL.Path) >= 7 && c.Request().URL.Path[0:7] == "/static" {
			return next(c)
		}
		// verifica se existe sessão
		usuario_id, err := verificaSessao(c)
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
	usuario_id, err := verificaSessao(c)
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
	// cria a struct desse endpoint
	ch := make(chan struct{})
	subs := Subscriber{
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
		case evento := <-ch:
			fmt.Println("Recebido evento em:", evento)
			// antes de enviar qualquer update, verifica se a sessão ainda está válida
			usuario_id, err := verificaSessao(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			// se não estiver válida, redireciona pra login
			if usuario_id == 0 {
				return sse.Redirect("/login")
			}

			// se estiver válida, manda update
			// pega as informações atualizadas das listas no banco
			// queries := models.New(db)
			// listas, err := queries.PegaListas(c.Request().Context(), usuario_id)
			// if err != nil {
			// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			// }

			// envia pro front com sse.MergeFragmentTempl()
			// vt := datastar.WithUseViewTransitions(true)
			// sse.MergeFragmentTempl(lc.ListaDeListas(listas), vt)
		}
	}
}

func handlerApiListaPost(c echo.Context) error {
	// pega usuario da sessao
	// usuario_id, err := verificaSessao(c)
	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// para manter separation os concerns, cria uma DTO pra receber e validar info do front
	listaPostDTO := struct {
		UsuarioID         int32    `json:"usuario_id"`
		Lista             string   `json:"lista" validate:"required"`
		Descricao         string   `json:"descricao"`
		Compartilhamentos []string `json:"compartilhamentos"`
	}{}
	// Unmarshella pra o DTO
	// err = c.Bind(&listaPostDTO)
	// if err != nil {
	// 	// TODO: trocar aqui pra validação dos campos
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// valida o DTO
	// instancia o validator
	validate := validator.New(validator.WithRequiredStructEnabled())
	// faz a validação
	err := validate.Struct(&listaPostDTO)
	// verifica se deu algo de errado
	if err != nil {
		var validateErrs validator.ValidationErrors
		// aqui, como só valida 1 campo, eu usei validateErrs[0], mas quando for mais, usar o for abaixo
		if errors.As(err, &validateErrs) && validateErrs[0].Tag() == "required" {
			sse := datastar.NewSSE(c.Response().Writer, c.Request())
			return sse.MergeSignals(
				[]byte(`{input_lista_erro: "Nome da lista é necessário..."}`),
			)
			// fmt.Println(validateErrs[0].Error())

			// for _, e := range validateErrs {
			// 	fmt.Println(e.Namespace())
			// 	fmt.Println(e.Field())
			// 	fmt.Println(e.StructNamespace())
			// 	fmt.Println(e.StructField())
			// 	fmt.Println(e.Tag())
			// 	fmt.Println(e.ActualTag())
			// 	fmt.Println(e.Kind())
			// 	fmt.Println(e.Type())
			// 	fmt.Println(e.Value())
			// 	fmt.Println(e.Param())
			// 	fmt.Println()
			// }
		}
	}

	// daqui pra frente, a informação recebida está validada e pronta pra ser inserida no banco

	// cria o objeto que vai pro banco com as informações validadas no DTO
	// lista := models.InsereListaParams(listaPostDTO)
	// lista := models.InsereListaParams{
	// 	UsuarioID: listaPostDTO.UsuarioID,
	// 	Lista:     listaPostDTO.Lista,
	// 	Descricao: listaPostDTO.Descricao,
	// 	//	Compartilhamentos: listaPostDTO.Compartilhamentos,
	// }
	// // acrescenta o ID do usuário
	// lista.UsuarioID = usuario_id

	// // insere no banco as informações recebidas
	// queries := models.New(db)
	// err = queries.InsereLista(c.Request().Context(), lista)
	// if err != nil {
	// 	var pgErr *pgconn.PgError
	// 	if errors.As(err, &pgErr) {
	// 		// fmt.Println(pgErr.Message) // => syntax error at end of input
	// 		// fmt.Println(pgErr.Code)    // => 42601
	// 		// error code para Lista duplicada
	// 		if pgErr.Code == "23505" {
	// 			sse := datastar.NewSSE(c.Response().Writer, c.Request())
	// 			erro := fmt.Sprintf(`{input_lista_erro: 'Lista %v já existe.'}`, lista.Lista)
	// 			return sse.MergeSignals([]byte(erro))
	// 		}
	// 	}

	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// limpa msg de erro caso exista
	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	limpa_erro := `{input_lista_erro: ''}`
	sse.MergeSignals([]byte(limpa_erro))

	// manda evento pra publicação
	publisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

func handlerApiListaDelete(c echo.Context) error {
	// var lista models.DeletaListaParams
	// pathParams := struct {
	// 	ListaID int32 `param:"id"`
	// }{}

	// pega usuario da sessao
	// usuario_id, err := verificaSessao(c)
	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// pega parâmetros do path
	// err = c.Bind(&pathParams)
	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// acrescenta usuario_id que pegou da sessão
	// lista.UsuarioID = usuario_id
	// lista.ListaID = pathParams.ListaID

	// faz o delete
	// queries := models.New(db)
	// err = queries.DeletaLista(c.Request().Context(), lista)
	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// se der certo a inclusão da lista no banco, manda evento pra publicação
	// tem que mandar evento para o path do delete sem o id recebido, ou seja,
	// /api/lista/11, por exemplo, tem que virar /api/lista
	// desmontaURL := strings.Split(c.Request().URL.Path, "/")
	// montaPathSemID := strings.Join(desmontaURL[:len(desmontaURL)-1], "/")

	// uma vez que o Path esteja montado corretamente, manda o evento
	publisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

func handlerApiListaPatch(c echo.Context) error {
	var listaAtualizada models.AtualizaListaParams

	SignalsAndParams := struct {
		ListaID        int32  `param:"id"`
		Lista          string `json:"lista_editada"`
		Descricao      string `json:"descricao_editada"`
		EditarLista    bool   `json:"editar_lista"`
		CancelarEdicao bool   `json:"cancelar_edicao"`
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

	// se o frontend pediu para editar ou cancelar a edição de uma lista
	if SignalsAndParams.EditarLista || SignalsAndParams.CancelarEdicao {
		// busca no banco a lista a ser alterada pra enviar pro front o componente preenchido
		// queries := models.New(db)
		// lista, err := queries.PegaLista(c.Request().Context(), SignalsAndParams.ListaID)
		// if err != nil {
		// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		// }

		// abre sse
		sse := datastar.NewSSE(c.Response().Writer, c.Request())

		// se é pra editar a lista, manda o fragmento pra edição
		if SignalsAndParams.EditarLista {
			// manda um componente de edição da lista
			// sse.MergeFragmentTempl(lc.EditarLista(lista))

			// retorna um merge de signals pra editar_lista = false
			return sse.MergeSignals([]byte(`{ editar_lista: false }`))
		}

		// se cancelou a edição, manda o fragmento de card de lista
		// converte a lista de uma struct pra outra, por serem identicas e só mudarem o nome, dá certo assim
		// listaConvertida := models.PegaListasRow(lista)
		// sse.MergeFragmentTempl(lc.Lista(listaConvertida))

		// retorna um merge de signals pra editar_lista = false
		return sse.MergeSignals([]byte(`{ cancelar_edicao: false }`))
	}

	// se chegou até aqui, faz o update

	// preenche a struct do banco
	listaAtualizada.UsuarioID = usuario_id
	listaAtualizada.ListaID = SignalsAndParams.ListaID
	listaAtualizada.Lista = SignalsAndParams.Lista
	if SignalsAndParams.Descricao != "" {
		listaAtualizada.Descricao.String = SignalsAndParams.Descricao
		listaAtualizada.Descricao.Valid = true
	}
	// manda pro banco
	// queries := models.New(db)
	// err = queries.AtualizaLista(c.Request().Context(), listaAtualizada)
	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	// uma vez que o Path esteja montado corretamente, manda o evento
	publisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

// TODO: migrar o daisyui/tailwind de CDN para o node pra conseguir usar o folha de modelo para o choicesjs
