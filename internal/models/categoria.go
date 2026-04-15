package models

import "github.com/jackc/pgx/v5/pgtype"

type Categoria struct {
	ID           int32              `json:"id"`
	UsuarioID    int32              `json:"usuario_id"`
	Categoria    string             `json:"categoria"`
	CriadoEm     pgtype.Timestamptz `json:"criado_em"`
	AtualizadoEm pgtype.Timestamptz `json:"atualizado_em"`
}
