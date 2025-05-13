package models

const (
	sqlSelectTestaLogin = `
		SELECT id, usuario
		FROM listas.usuarios
		WHERE usuario = $1 AND senha = $2
		LIMIT 1
		;
	`
	sqlSelectUsuariosParaCompartilhar = `
		SELECT 
			cast(u.id as TEXT) as id, 
			u.nome
		FROM listas.usuarios u
		WHERE u.id != $1
		ORDER BY u.nome
		;
	`
)
