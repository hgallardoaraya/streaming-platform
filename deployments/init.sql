-- Crear otra base de datos
CREATE DATABASE generaldb;

-- Usar la base de datos recién creada
\c generaldb;

-- Crear una tabla en la nueva base de datos con la combinación de stream_name y username única
CREATE TABLE stream (
    id SERIAL PRIMARY KEY,
    stream_key VARCHAR(100) NOT NULL,
    stream_name VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_stream_name_username UNIQUE (stream_name, username)  -- Combinación única
);
