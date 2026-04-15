package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"github.com/vinifrancosilva/lista_v2/internal/models"
	"github.com/vinifrancosilva/lista_v2/internal/utils"
	lc "github.com/vinifrancosilva/lista_v2/views/components/listas"
)

// ----------------------- API -----------------------

// Lista
type HandlerLista struct {
	PubSub *models.PubSubChanels
}

func NewHandlerLista(pb *models.PubSubChanels) *HandlerLista {
	return &HandlerLista{
		PubSub: pb,
	}
}

// Get - abre o canal SSE que recebe todas as mudancas realizadas na lista
func (h *HandlerLista) ListaGet(c echo.Context) error {
	// pega usuario da sessao
	usuario, err := utils.VerificaSessao(c)
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
	h.PubSub.SubscriberChan <- subs

	for {
		select {
		// se fechar a conexão, faz o unsubscribe
		case <-c.Request().Context().Done():
			// fecha o canal
			h.PubSub.UnsubscriberChan <- subs
			return nil
		case <-ch:
			// antes de enviar qualquer update, verifica se a sessão ainda está válida
			usuario_check, err := utils.VerificaSessao(c)
			if err != nil {
				h.PubSub.UnsubscriberChan <- subs
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			// se não estiver válida, redireciona pra login
			if usuario_check.ID == 0 {
				h.PubSub.UnsubscriberChan <- subs
				sse.Redirect("/login")
				return nil
			}

			// se estiver válida, pega as listas atualizadas
			listas, err := models.PegaListas(c.Request().Context(), &usuario)
			if err != nil {
				h.PubSub.UnsubscriberChan <- subs
				sse.Redirect("/login")
				return nil
			}

			// envia pro front com sse.MergeFragmentTempl()
			vt := datastar.WithUseViewTransitions(true)
			sse.PatchElementTempl(lc.ListaDeListas(listas), vt)
		}
	}
}

func (h *HandlerLista) ListaCreatePost(c echo.Context) error {
	// pega usuario da sessao
	usuario, err := utils.VerificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// para manter separation os concerns, cria uma DTO pra receber e validar info do front
	listaPostDTO := struct {
		Lista string `json:"lista" validate:"required"`
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
			return sse.MarshalAndPatchSignals(map[string]any{
				"input_lista_erro": "Nome da lista é necessário...",
			})
		}
	}

	// insere no banco as informações validadas
	err = models.InsereLista(c.Request().Context(), listaPostDTO.Lista, &usuario)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// limpa msg de erro caso exista
	sse := datastar.NewSSE(c.Response().Writer, c.Request())
	limpa_erro := `{input_lista_erro: ''}`
	sse.PatchSignals([]byte(limpa_erro))

	// manda evento pra publicação
	h.PubSub.PublisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

func (h *HandlerLista) ApiListaDelete(c echo.Context) error {
	// pega usuario da sessao
	usuario, err := utils.VerificaSessao(c)
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
	err = models.DeletaLista(c.Request().Context(), listaIDInt32, &usuario)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// manda evento pra publicação
	h.PubSub.PublisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}

func (h *HandlerLista) ApiListaPatch(c echo.Context) error {
	SignalsAndParams := struct {
		ListaID   int32  `param:"id"`
		Lista     string `json:"lista"`
		Descricao string `json:"descricao"`
	}{}

	// pega usuario da sessao
	usuario, err := utils.VerificaSessao(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// pega parâmetros do path
	err = c.Bind(&SignalsAndParams)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// cria uma struct do validator pra validar os dados recebidos
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(&SignalsAndParams)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			sse := datastar.NewSSE(c.Response().Writer, c.Request())
			return sse.MarshalAndPatchSignals(map[string]any{
				"input_lista_erro": "Nome da lista é necessário...",
			})
		}
	}

	lista := models.Lista{
		ID:        SignalsAndParams.ListaID,
		Lista:     SignalsAndParams.Lista,
		Descricao: pgtype.Text{String: SignalsAndParams.Descricao, Valid: true},
	}

	// faz o update
	err = models.AtualizaLista(c.Request().Context(), &lista, &usuario)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// manda evento pra publicação
	h.PubSub.PublisherChan <- "listas"

	return c.NoContent(http.StatusOK)
}
