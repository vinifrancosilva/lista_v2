package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Usuario struct {
	ID           int                `json:"id" db:"id"`
	Usuario      string             `json:"usuario" db:"usuario"`
	Senha        string             `json:"senha" db:"senha"`
	Nome         pgtype.Text        `json:"nome" db:"nome"`
	CriadoEm     pgtype.Timestamptz `json:"criado_em" db:"criado_em"`
	AtualizadoEm pgtype.Timestamptz `json:"atualizado_em" db:"atualizado_em"`
}

func (u *Usuario) TestaLogin(ctx context.Context) error {
	var usuario Usuario

	// err := db.GetContext(ctx, &usuario, sqlSelectTestaLogin, u.Usuario, u.Senha)
	err := db.GetContext(ctx, &usuario, sqlSelectTestaLogin, u.Usuario, u.Senha)
	if err != nil {
		if err == pgx.ErrNoRows {
			return err
		}
		return fmt.Errorf("erro ao buscar usuário: %w", err)
	}

	// Aqui você pode copiar os dados encontrados para o struct original, se quiser
	*u = usuario

	return nil
}
