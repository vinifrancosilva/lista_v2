package models

import "github.com/jackc/pgx/v5/pgtype"

type Compartilhamento struct {
	ID        int32              `json:"id"`
	ListaID   int32              `json:"lista_id"`
	UsuarioID int32              `json:"usuario_id"`
	CriadoEm  pgtype.Timestamptz `json:"criado_em"`
}
