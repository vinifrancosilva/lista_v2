-- name: TestaLogin :one
SELECT id, usuario
FROM listas.usuarios
WHERE usuario = $1 AND senha = $2
LIMIT 1;

-- name: ListaUsuarios :many
SELECT id, usuario FROM listas.usuarios
ORDER BY usuario;