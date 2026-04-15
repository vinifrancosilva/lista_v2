package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/vinifrancosilva/lista_v2/internal/handlers"
	"github.com/vinifrancosilva/lista_v2/internal/models"
)

func DefineRotas(e *echo.Echo, pb *models.PubSubChanels) {
	// Static
	e.Static("/static", "static")

	// Index
	e.GET("/", handlers.HandlerIndex)

	// Login / Logout
	e.GET("/login", handlers.HandlerLoginPage)
	e.POST("/login", handlers.HandlerLoginPost)
	e.GET("/logout", handlers.HandlerLogout)

	// Categorias
	e.GET("/categorias", handlers.HandlerFake)

	// API Group
	g := e.Group("/api")

	// API da Lista
	HandlerLista := handlers.NewHandlerLista(pb)
	g.GET("/lista", HandlerLista.ListaGet)
	g.POST("/lista/create", HandlerLista.ListaCreatePost)
	g.PATCH("/lista/:id", HandlerLista.ApiListaPatch)
	g.DELETE("/lista/:id", HandlerLista.ListaDelete)
}
