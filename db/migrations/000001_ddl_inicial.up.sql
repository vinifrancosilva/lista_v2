CREATE SCHEMA IF NOT EXISTS listas
    AUTHORIZATION vini
;

CREATE TABLE IF NOT EXISTS listas.usuarios (
	id          	  SERIAL PRIMARY KEY,
	usuario			    TEXT NOT NULL UNIQUE,
	senha			      TEXT NOT NULL,
	nome			      TEXT NULL,
	criado_em 		  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  atualizado_em	  TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
;

INSERT INTO listas.usuarios (usuario, senha, nome)
VALUES ('vini', 'Vini1406@', 'Papai')
;

CREATE TABLE IF NOT EXISTS listas.listas (
  id          	SERIAL PRIMARY KEY,
  usuario_id		INT NOT NULL,
  lista			    TEXT NOT NULL,
  descricao		  TEXT NULL,
  criado_em 		TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  atualizado_em	TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (usuario_id) REFERENCES listas.usuarios(id),
  UNIQUE (usuario_id, lista)
)
;

CREATE TABLE IF NOT EXISTS listas.categorias (
  id          	SERIAL PRIMARY KEY,
  usuario_id		INT NOT NULL,
  categoria		  TEXT NOT NULL,
  criado_em 		TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  atualizado_em	TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (usuario_id) REFERENCES listas.usuarios(id)
)
;

CREATE TABLE IF NOT EXISTS listas.items (
  id          	SERIAL PRIMARY KEY,
  lista_id		  INT NOT NULL,
  categoria_id	INT NULL,
  item		    	TEXT NOT NULL,
  feito         BOOLEAN NOT NULL DEFAULT FALSE,
  criado_em 		TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  atualizado_em	TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (lista_id) REFERENCES listas.listas(id),
  FOREIGN KEY (categoria_id) REFERENCES listas.categorias(id)
)
;

CREATE TABLE IF NOT EXISTS listas.compartilhamentos (
  id          	SERIAL PRIMARY KEY,
  lista_id		  INT NOT NULL,
  usuario_id		INT NOT NULL,
  criado_em 		TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY   (lista_id) REFERENCES listas.listas(id),
  FOREIGN KEY   (usuario_id) REFERENCES listas.usuarios(id)
)
;
