-- 1. Criação do Banco (Executado no banco padrão 'postgres')
DROP DATABASE IF EXISTS vinifranco;

CREATE DATABASE vinifranco
    WITH
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8'
    LOCALE_PROVIDER = 'libc'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1
    IS_TEMPLATE = False;

-- 2. Mudança de Conexão
-- Se estiver usando o terminal (psql), use o comando abaixo:
\c vinifranco

-- 3. Criação de Usuário e Permissões de Banco
-- (Usuários são globais no cluster, então a ordem aqui é flexível)
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'vini') THEN
        CREATE USER vini WITH PASSWORD 'Trader1406Bauru';
    END IF;
END
$$;

GRANT ALL ON DATABASE vinifranco TO postgres;
GRANT ALL ON DATABASE vinifranco TO vini;
GRANT TEMPORARY, CONNECT ON DATABASE vinifranco TO PUBLIC;

-- 4. Criação do Schema (Agora dentro do banco vinifranco)
CREATE SCHEMA IF NOT EXISTS listas AUTHORIZATION vini;

-- Garantir que o vini tenha acesso total ao schema que ele é dono
GRANT ALL ON SCHEMA listas TO vini;
