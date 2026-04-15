package models

import "github.com/jackc/pgx/v5/pgtype"

type Item struct {
	ID           int                `json:"id"`
	ListaID      int                `json:"lista_id"`
	CategoriaID  pgtype.Int4        `json:"categoria_id"`
	Item         string             `json:"item"`
	Feito        bool               `json:"feito"`
	CriadoEm     pgtype.Timestamptz `json:"criado_em"`
	AtualizadoEm pgtype.Timestamptz `json:"atualizado_em"`
}
