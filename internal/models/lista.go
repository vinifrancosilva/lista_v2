package models

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
)

// PegaListasRow é o retorno da query que pega as listas do usuário
type PegaListasRow struct {
	ID        int32       `json:"id" db:"id"`
	Lista     string      `json:"lista" db:"lista"`
	Descricao pgtype.Text `json:"descricao" db:"descricao"`
}

// PegaListas retorna todas as listas do usuário
func (u *Usuario) PegaListas(ctx context.Context) ([]PegaListasRow, error) {
	var listas []PegaListasRow
	err := GetDB().SelectContext(ctx, &listas, `
		SELECT id, lista, descricao 
		FROM listas.listas
		WHERE usuario_id = $1
		ORDER BY criado_em DESC
	`, u.ID)
	return listas, err
}

// PegaLista retorna uma lista específica
func PegaLista(ctx context.Context, listaID, usuarioID int32) (*Lista, error) {
	var lista Lista
	err := GetDB().GetContext(ctx, &lista, `
		SELECT id, usuario_id, lista, descricao, criado_em, atualizado_em
		FROM listas.listas
		WHERE id = $1 AND usuario_id = $2
	`, listaID, usuarioID)
	if err != nil {
		return nil, err
	}
	return &lista, nil
}

// InsereLista insere uma nova lista
func (u *Usuario) InsereLista(ctx context.Context, listaName, descricao string) error {
	_, err := GetDB().ExecContext(ctx, `
		INSERT INTO listas.listas (usuario_id, lista, descricao)
		VALUES ($1, $2, $3)
	`, u.ID, listaName, descricao)
	return err
}

// AtualizaLista atualiza uma lista existente
func AtualizaLista(ctx context.Context, listaID, usuarioID int32, listaName, descricao string) error {
	_, err := GetDB().ExecContext(ctx, `
		UPDATE listas.listas SET
			lista = $1,
			descricao = $2,
			atualizado_em = NOW()
		WHERE id = $3 AND usuario_id = $4
	`, listaName, descricao, listaID, usuarioID)
	return err
}

// DeletaLista deleta uma lista
func DeletaLista(ctx context.Context, listaID, usuarioID int32) error {
	_, err := GetDB().ExecContext(ctx, `
		DELETE FROM listas.listas
		WHERE id = $1 AND usuario_id = $2
	`, listaID, usuarioID)
	return err
}
