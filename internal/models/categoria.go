package models

import "github.com/jackc/pgx/v5/pgtype"

type Categoria struct {
	ID           int                `json:"id" db:"id"`
	UsuarioID    int                `json:"usuario_id" db:"usuario_id"`
	Categoria    string             `json:"categoria" db:"categoria"`
	CriadoEm     pgtype.Timestamptz `json:"criado_em" db:"criado_em"`
	AtualizadoEm pgtype.Timestamptz `json:"atualizado_em" db:"atualizado_em"`
}
