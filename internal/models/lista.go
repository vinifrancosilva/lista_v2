package models

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrListaDuplicada = errors.New("lista com esse nome já existe para este usuário")

type Lista struct {
	ID           int                `param:"id" json:"id" db:"id"`
	UsuarioID    int                `json:"usuario_id" db:"usuario_id"`
	Lista        string             `json:"lista" db:"lista"`
	Descricao    pgtype.Text        `json:"descricao" db:"descricao"`
	Quantidade   int                `json:"quantidade" db:"quantidade"`
	CriadoEm     pgtype.Timestamptz `json:"criado_em" db:"criado_em"`
	AtualizadoEm pgtype.Timestamptz `json:"atualizado_em" db:"atualizado_em"`
}

// ListaDbSelect é o retorno da query que pega as listas do usuário
type ListaDbSelect struct {
	ID         int         `json:"id" db:"id"`
	UsuarioID  int         `json:"usuario_id" db:"usuario_id"`
	Lista      string      `json:"lista" db:"lista"`
	Descricao  pgtype.Text `json:"descricao" db:"descricao"`
	Total      int         `json:"total" db:"total"`
	Concluidos int         `json:"concluidos" db:"concluidos"`
}

type ListaEdicao struct {
	Lista             Lista             `json:"lista"`
	Categorias        []CategoriaEdicao `json:"categorias"`
	Compartilhamentos []UsuarioEdicao   `json:"compartilhamentos"`
}

type CategoriaEdicao struct {
	Categoria Categoria `json:"categoria"`
	Vinculada bool      `json:"vinculada"`
}

type UsuarioEdicao struct {
	Usuario       Usuario `json:"usuario"`
	Compartilhado bool    `json:"compartilhado"`
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
func PegaListaEdicao(ctx context.Context, lista *Lista, usuario *Usuario) (*ListaEdicao, error) {
	var listaEdicao ListaEdicao

	// fazendo a query retornar json pra fazer apenas 1 query ao invés de 3
	// (1 pra lista, 1 pras categorias e 1 pros compartilhamentos)
	var raw json.RawMessage
	err := db.GetContext(ctx, &raw, `
		SELECT json_build_object(
			'lista', json_build_object(
				'id', l.id,
				'usuario_id', l.usuario_id,
				'lista', l.lista,
				'descricao', l.descricao,
				'quantidade', COALESCE((SELECT COUNT(*) FROM listas.items i WHERE i.lista_id = l.id), 0),
				'criado_em', l.criado_em,
				'atualizado_em', l.atualizado_em
			),
			'categorias', COALESCE((
				SELECT json_agg(json_build_object(
					'categoria', json_build_object(
						'id', c.id,
						'usuario_id', c.usuario_id,
						'categoria', c.categoria,
						'criado_em', c.criado_em,
						'atualizado_em', c.atualizado_em
					),
					'vinculada', EXISTS (SELECT 1 FROM listas.items i WHERE i.categoria_id = c.id AND i.lista_id = $1)
				) ORDER BY c.categoria)
				FROM listas.categorias c
				WHERE c.usuario_id = $2
			), '[]'::json),
			'compartilhamentos', COALESCE((
				SELECT json_agg(json_build_object(
					'usuario', json_build_object(
						'id', u.id,
						'usuario', u.usuario,
						'nome', u.nome,
						'criado_em', u.criado_em,
						'atualizado_em', u.atualizado_em
					),
					'compartilhado', CASE WHEN ch.usuario_id IS NOT NULL THEN true ELSE false END
				) ORDER BY u.nome)
				FROM listas.usuarios u
				LEFT JOIN listas.compartilhamentos ch ON ch.usuario_id = u.id AND ch.lista_id = $1
				WHERE u.id != $2
			), '[]'::json)
		)
		FROM listas.listas l
		WHERE l.id = $1 AND l.usuario_id = $2
	`, lista.ID, usuario.ID)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(raw, &listaEdicao); err != nil {
		return nil, err
	}

	return &listaEdicao, nil
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
	_, err := db.ExecContext(ctx, `
		UPDATE listas.listas SET
			lista = $1,
			descricao = $2,
			atualizado_em = NOW()
		WHERE id = $3 AND usuario_id = $4
	`, lista.Lista, lista.Descricao, lista.ID, usuario.ID)
	return err
}

// DeletaLista deleta uma lista
func DeletaLista(ctx context.Context, lista *Lista, usuario *Usuario) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM listas.listas
		WHERE id = $1 AND usuario_id = $2
	`, lista.ID, usuario.ID)
	return err
}
