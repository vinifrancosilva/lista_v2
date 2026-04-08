package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vinifrancosilva/lista_v2/internal/models"
	"github.com/vinifrancosilva/lista_v2/internal/utils"
	"github.com/vinifrancosilva/lista_v2/views/pages"
)

//[ ] TODO: ao compartilhar a lista com outro usuario, compartilha também as categorias

// Handlers dos endpoints
func HandlerIndex(c echo.Context) error {
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

// TODO: migrar o daisyui/tailwind de CDN para o node pra conseguir usar o folha de modelo para o choicesjs
