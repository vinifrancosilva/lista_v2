package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// const atualizaLista = `-- name: AtualizaLista :exec
// UPDATE listas.listas SET
//
//	lista = $1,
//	descricao = $2,
//	atualizado_em = NOW()
//
// WHERE 1 = 1
//
//	AND id = $3
//	AND usuario_id = $4
//
// `
type Queries struct {
	db *pgx.Conn
}

func NewQueries(db *pgx.Conn) *Queries {
	return &Queries{db: db}
}

type AtualizaListaParams struct {
	Lista     string      `json:"lista"`
	Descricao pgtype.Text `json:"descricao"`
	ListaID   int         `json:"lista_id"`
	UsuarioID int         `json:"usuario_id"`
}

// func (q *Queries) AtualizaLista(ctx context.Context, arg AtualizaListaParams) error {
// 	_, err := q.db.Exec(ctx, atualizaLista,
// 		arg.Lista,
// 		arg.Descricao,
// 		arg.ListaID,
// 		arg.UsuarioID,
// 	)
// 	return err
// }

// const deletaLista = `-- name: DeletaLista :exec
// DELETE FROM listas.listas
// WHERE 1 = 1
//   AND usuario_id = $1
//   AND id = $2
// `

type DeletaListaParams struct {
	UsuarioID int `json:"usuario_id"`
	ListaID   int `json:"lista_id"`
}

// func DeletaLista(ctx context.Context, arg DeletaListaParams) error {
// 	_, err := db.Exec(ctx, deletaLista, arg.UsuarioID, arg.ListaID)
// 	return err
// }

type InsereListaParams struct {
	UsuarioID int    `json:"usuario_id"`
	Lista     string `json:"lista"`
	Descricao string `json:"descricao"`
}

const insereLista = `
	INSERT INTO listas.listas (
		usuario_id,
		lista,
		descricao
	) VALUES (
		$1,
		$2::text,
		$3::text
	)
`

func (q *Queries) InsereLista(ctx context.Context, arg InsereListaParams) error {
	_, err := q.db.Exec(ctx, insereLista, arg.UsuarioID, arg.Lista, arg.Descricao)
	return err
}

// const listaUsuarios = `-- name: ListaUsuarios :many
// SELECT id, usuario FROM listas.usuarios
// ORDER BY usuario
// `

type UsuarioDbSelect struct {
	ID      int    `json:"id"`
	Usuario string `json:"usuario"`
}

// func (q *Queries) ListaUsuarios(ctx context.Context) ([]ListaUsuariosRow, error) {
// 	rows, err := q.db.Query(ctx, listaUsuarios)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	var items []ListaUsuariosRow
// 	for rows.Next() {
// 		var i ListaUsuariosRow
// 		if err := rows.Scan(&i.ID, &i.Usuario); err != nil {
// 			return nil, err
// 		}
// 		items = append(items, i)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}
// 	return items, nil
// }

// const pegaLista = `-- name: PegaLista :one
// SELECT id, lista, descricao FROM listas.listas
// WHERE id = $1
// `

// func (q *Queries) PegaLista(ctx context.Context, listaID int) (PegaListaRow, error) {
// 	row := q.db.QueryRow(ctx, pegaLista, listaID)
// 	var i PegaListaRow
// 	err := row.Scan(&i.ID, &i.Lista, &i.Descricao)
// 	return i, err
// }

// const pegaListas = `-- name: PegaListas :many
// SELECT id, lista, descricao FROM listas.listas
// WHERE usuario_id = $1
// ORDER BY criado_em DESC
// `

// func (q *Queries) PegaListas(ctx context.Context, usuarioID int) ([]PegaListasRow, error) {
// 	rows, err := q.db.Query(ctx, pegaListas, usuarioID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	var items []PegaListasRow
// 	for rows.Next() {
// 		var i PegaListasRow
// 		if err := rows.Scan(&i.ID, &i.Lista, &i.Descricao); err != nil {
// 			return nil, err
// 		}
// 		items = append(items, i)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}
// 	return items, nil
// }

// const pegaUsuariosCompartilhamento = `-- name: PegaUsuariosCompartilhamento :many
// SELECT
//   cast(u.id as TEXT) as id,
//   u.nome
// FROM listas.usuarios u
// WHERE u.id != $1
// ORDER BY u.nome
// `

type PegaUsuariosCompartilhamentoRow struct {
	ID   string      `json:"id"`
	Nome pgtype.Text `json:"nome"`
}

// func (q *Queries) PegaUsuariosCompartilhamento(ctx context.Context, usuarioID int) ([]PegaUsuariosCompartilhamentoRow, error) {
// 	rows, err := q.db.Query(ctx, pegaUsuariosCompartilhamento, usuarioID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	var items []PegaUsuariosCompartilhamentoRow
// 	for rows.Next() {
// 		var i PegaUsuariosCompartilhamentoRow
// 		if err := rows.Scan(&i.ID, &i.Nome); err != nil {
// 			return nil, err
// 		}
// 		items = append(items, i)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}
// 	return items, nil
// }

// const testaLogin = `-- name: TestaLogin :one
// SELECT id, usuario
// FROM listas.usuarios
// WHERE usuario = $1 AND senha = $2
// LIMIT 1
// `

type TestaLoginParams struct {
	Usuario string `json:"usuario"`
	Senha   string `json:"senha"`
}

type LoginRow struct {
	ID      int    `json:"id"`
	Usuario string `json:"usuario"`
}

func (u *Usuario) PegaUsuariosParaCompartilhar(ctx context.Context) ([]Usuario, error) {
	var usuarios []Usuario

	err := db.SelectContext(ctx, &usuarios, sqlSelectUsuariosParaCompartilhar, u.ID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar usuários para compartilhar: %w", err)
	}

	return usuarios, nil
}
