package models

import (
	"github.com/jmoiron/sqlx"
)

var (
	db *sqlx.DB
)

// DBInit inicializa o banco de dados e retorna a conexão
func DBInit(dbUrl string) error {
	var err error

	db, err = sqlx.Connect("pgx", dbUrl)
	if err != nil {
		return err
	}

	return nil
}
