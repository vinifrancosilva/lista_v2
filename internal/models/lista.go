package models

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrListaDuplicada = errors.New("lista com esse nome já existe para este usuário")

type Lista struct {
	ID           int32              `json:"id" db:"id"`
	UsuarioID    int32              `json:"usuario_id" db:"usuario_id"`
	Lista        string             `json:"lista" db:"lista"`
	Descricao    pgtype.Text        `json:"descricao" db:"descricao"`
	Quantidade   int32              `json:"quantidade" db:"quantidade"`
	CriadoEm     pgtype.Timestamptz `json:"criado_em" db:"criado_em"`
	AtualizadoEm pgtype.Timestamptz `json:"atualizado_em" db:"atualizado_em"`
}

// ListaDbSelect é o retorno da query que pega as listas do usuário
type ListaDbSelect struct {
	ID         int32       `json:"id" db:"id"`
	UsuarioID  int32       `json:"usuario_id" db:"usuario_id"`
	Lista      string      `json:"lista" db:"lista"`
	Descricao  pgtype.Text `json:"descricao" db:"descricao"`
	Total      int32       `json:"total" db:"total"`
	Concluidos int32       `json:"concluidos" db:"concluidos"`
}

// PegaListas retorna todas as listas do usuário (proprietário e compartilhadas)
func PegaListas(ctx context.Context, u *Usuario) ([]ListaDbSelect, error) {
	var listas []ListaDbSelect
	err := db.SelectContext(ctx, &listas, `
		SELECT 
			l.id,
			l.usuario_id,
			l.lista,
			l.descricao,
			COALESCE(COUNT(i.id), 0) as total,
			COALESCE(SUM(CASE WHEN i.feito = true THEN 1 ELSE 0 END), 0) as concluidos
		FROM listas.listas l
		LEFT JOIN listas.items i ON l.id = i.lista_id
		WHERE l.usuario_id = $1 
			OR l.id IN (
				SELECT lista_id FROM listas.compartilhamentos 
				WHERE usuario_id = $1
			)
		GROUP BY l.id, l.usuario_id, l.lista, l.descricao
		ORDER BY l.criado_em DESC
	`,
		u.ID,
	)
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
func InsereLista(ctx context.Context, listaNome string, u *Usuario) error {
	_, err := db.ExecContext(
		ctx,
		`
		INSERT INTO listas.listas (usuario_id, lista)
		VALUES ($1, $2)
	`,
		u.ID,
		listaNome,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrListaDuplicada
		}
	}
	return err
}

// AtualizaLista atualiza uma lista existente
func AtualizaLista(ctx context.Context, lista *Lista, usuario *Usuario) error {
	_, err := GetDB().ExecContext(ctx, `
		UPDATE listas.listas SET
			lista = $1,
			descricao = $2,
			atualizado_em = NOW()
		WHERE id = $3 AND usuario_id = $4
	`, lista.Lista, lista.Descricao, lista.ID, usuario.ID)
	return err
}

// DeletaLista deleta uma lista
func DeletaLista(ctx context.Context, listaID int32, usuario *Usuario) error {
	_, err := GetDB().ExecContext(ctx, `
		DELETE FROM listas.listas
		WHERE id = $1 AND usuario_id = $2
	`, listaID, usuario.ID)
	return err
}
