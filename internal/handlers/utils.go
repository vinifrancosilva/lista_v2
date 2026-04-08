package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

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
func HandlerFake(c echo.Context) error {
	return c.String(http.StatusOK, "FAKE HANDLER")
}
