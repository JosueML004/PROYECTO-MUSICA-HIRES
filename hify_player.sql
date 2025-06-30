-- Crear base de datos
CREATE DATABASE IF NOT EXISTS hify_player;
USE hify_player;

-- Tabla de usuarios para registro e inicio de sesión
CREATE TABLE IF NOT EXISTS usuarios (
    id INT AUTO_INCREMENT PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    clave VARCHAR(255) NOT NULL,
    activo BOOLEAN DEFAULT FALSE
);

-- Tabla para almacenar metadatos de archivos de música .flac
CREATE TABLE IF NOT EXISTS musicas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    artist VARCHAR(255),
    album VARCHAR(255),
    title VARCHAR(255),
    genre VARCHAR(255),
    duration BIGINT,
    path VARCHAR(500) UNIQUE
);

-- Índices para optimizar búsquedas
CREATE INDEX idx_usuarios_email ON usuarios(email);
CREATE INDEX idx_musicas_artist ON musicas(artist);
CREATE INDEX idx_musicas_album ON musicas(album);
CREATE INDEX idx_musicas_title ON musicas(title);
CREATE INDEX idx_musicas_genre ON musicas(genre);

-- Credenciales para conexión a la base de datos
-- Usuario: root
-- Clave: 03042008
-- Host: localhost
-- Puerto: 3306
