-- name: TestaLogin :one
SELECT id, usuario
FROM listas.usuarios
WHERE usuario = $1 AND senha = $2
LIMIT 1
;

-- name: ListaUsuarios :many
SELECT id, usuario FROM listas.usuarios
ORDER BY usuario
;

-- name: PegaListas :many
SELECT id, lista, descricao FROM listas.listas
WHERE usuario_id = @usuario_id
ORDER BY criado_em DESC
;

-- name: PegaLista :one
SELECT id, lista, descricao FROM listas.listas
WHERE id = @lista_id
;

-- name: InsereLista :exec
INSERT INTO listas.listas (
  usuario_id,
  lista,
  descricao
) VALUES (
  @usuario_id,
  @lista::text,
  @descricao::text
)
;

-- name: DeletaLista :exec
DELETE FROM listas.listas
WHERE 1 = 1
  AND usuario_id = @usuario_id
  AND id = @lista_id
;

-- name: AtualizaLista :exec
UPDATE listas.listas SET
  lista = @lista,
  descricao = @descricao,
  atualizado_em = NOW()
WHERE 1 = 1
  AND id = @lista_id
  AND usuario_id = @usuario_id
;

-- name: PegaUsuariosCompartilhamento :many
SELECT 
  cast(u.id as TEXT) as id, 
  u.nome
FROM listas.usuarios u
WHERE u.id != @usuario_id
ORDER BY u.nome
;