package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/antonlindstrom/pgstore"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vinifrancosilva/lista_v2/config"
	"github.com/vinifrancosilva/lista_v2/internal/handlers"
	"github.com/vinifrancosilva/lista_v2/internal/models"
	"github.com/vinifrancosilva/lista_v2/internal/pubsub"
	"github.com/vinifrancosilva/lista_v2/internal/routes"
)

/*
TODOS:
x Rodar o postgres
x Implementar a session pelo banco
x Configurar o sqlc
- Acertar o templ
- Configurar o datastar sdk e pelo sdk chamar o latest datastar disponivel
- Configurar o tailwindcss com daisyui
- Configurar melhor o taskfile.yml para o meu setup
-
*/
var (
	// db             *sqlx.DB
	pgSessionStore *pgstore.PGStore
)

func main() {
	// Inicializa o banco de dados
	// db, pgSessionStore = dbInit()
	pgSessionStore = dbInit()
	// defer db.Close()
	defer pgSessionStore.Close()

	// Run a background goroutine to clean up expired sessions from the database.
	defer pgSessionStore.StopCleanup(pgSessionStore.Cleanup(time.Minute * 5))

	// Echo instance
	e := echo.New()

	// Cria os PubSub Channels
	pb := models.NewPubSubChannels()

	// Middlewares
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins:     []string{"http://localhost:8888", "http://localhost:3000", "http://localhost:5173"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
	// 	AllowCredentials: true,
	// }))
	e.Use(session.Middleware(pgSessionStore))
	e.Use(handlers.MiddlewareEstaLogado)

	// Routes
	routes.DefineRotas(e, &pb)

	// Roda o controle de conexões SSE
	go pubsub.ControleConexoesSSE(&pb)
	// go testaControleConexoesSSE()

	// Start server
	if err := e.Start(":8888"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

func dbInit() *pgstore.PGStore {
	// Gera as configurações do app a partir das variáveis de ambiente
	appConfig := config.AppConfig{
		DbUser:           os.Getenv("DB_USER"),
		DbPassword:       os.Getenv("DB_PASSWORD"),
		DbHost:           os.Getenv("DB_HOST"),
		DbPort:           os.Getenv("DB_PORT"),
		DbName:           os.Getenv("DB_NAME"),
		DbSSLMode:        os.Getenv("DB_SSLMODE"),
		DbSearchPath:     os.Getenv("DB_SEARCH_PATH"),
		SessionSecretKey: os.Getenv("SESSION_SECRET_KEY"),
	}

	// Cria a string de conexão com o banco de dados
	// searchPath := appConfig.DbSearchPath
	// if searchPath == "" {
	// 	searchPath = "public"
	// }
	// options := url.QueryEscape(fmt.Sprintf("-c search_path=%s", searchPath))

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
		// "postgres://%s:%s@%s:%s/%s?sslmode=%s&options=%s",
		appConfig.DbUser,
		appConfig.DbPassword,
		appConfig.DbHost,
		appConfig.DbPort,
		appConfig.DbName,
		appConfig.DbSSLMode,
		appConfig.DbSearchPath,
		// options,
	)

	// Cria uma store
	store, err := pgstore.NewPGStore(dbURL, appConfig.SessionSecretKeyByte())
	if err != nil {
		log.Fatalf("falha na criação da session store: %v", err)
	}

	err = models.DBInit(dbURL)
	if err != nil {
		log.Fatalf("falha na conexão com o PostgreSQL: %v", err)
	}

	return store
}

func testaControleConexoesSSE(pb *models.PubSubChanels) {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		pb.PublisherChan <- "/api/lista"
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("erro ao ler arquivo .env: %v", err)
	}
}
