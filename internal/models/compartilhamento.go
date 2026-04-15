package models

import "github.com/jackc/pgx/v5/pgtype"

type Compartilhamento struct {
	ID        int                `json:"id" db:"id"`
	ListaID   int                `json:"lista_id" db:"lista_id"`
	UsuarioID int                `json:"usuario_id" db:"usuario_id"`
	CriadoEm  pgtype.Timestamptz `json:"criado_em" db:"criado_em"`
}
