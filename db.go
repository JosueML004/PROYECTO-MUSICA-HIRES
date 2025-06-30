package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	conn *sql.DB
}

type Usuario struct {
	ID     int
	Nombre string
	Email  string
	Clave  string
	Activo bool
}

func NewDB(user, password, host, dbname string) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", user, password, host, dbname)
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = conn.Ping()
	if err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func (db *DB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}

// Crear tabla usuarios si no existe
func (db *DB) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id INT AUTO_INCREMENT PRIMARY KEY,
		nombre VARCHAR(100) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		clave VARCHAR(255) NOT NULL,
		activo BOOLEAN DEFAULT FALSE
	);`
	_, err := db.conn.Exec(query)
	if err != nil {
		return err
	}
	// Crear tabla musicas si no existe
	queryMusic := `
	CREATE TABLE IF NOT EXISTS musicas (
		id INT AUTO_INCREMENT PRIMARY KEY,
		artist VARCHAR(255),
		album VARCHAR(255),
		title VARCHAR(255),
		duration BIGINT,
		path VARCHAR(500) UNIQUE
	);`
	_, err = db.conn.Exec(queryMusic)
	return err
}

// Registrar usuario con contraseña hasheada
func (db *DB) AddUser(nombre, email, clave string) error {
	// Verificar si usuario existe
	var exists bool
	err := db.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE email=?)", email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("el usuario ya existe")
	}

	// Hashear contraseña
	hashed, err := bcrypt.GenerateFromPassword([]byte(clave), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.conn.Exec("INSERT INTO usuarios (nombre, email, clave) VALUES (?, ?, ?)", nombre, email, string(hashed))
	return err
}

func (db *DB) AddOrUpdateMusic(m MusicFile) error {
	// Verificar si la música existe por path
	var exists bool
	err := db.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM musicas WHERE path=?)", m.Path).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Actualizar registro existente
		_, err = db.conn.Exec("UPDATE musicas SET artist=?, album=?, title=?, duration=? WHERE path=?", m.Artist, m.Album, m.Title, int64(m.Duration.Seconds()), m.Path)
	} else {
		// Insertar nuevo registro
		_, err = db.conn.Exec("INSERT INTO musicas (artist, album, title, duration, path) VALUES (?, ?, ?, ?, ?)", m.Artist, m.Album, m.Title, int64(m.Duration.Seconds()), m.Path)
	}
	return err
}

func (db *DB) GetMusicList() ([]MusicFile, error) {
	rows, err := db.conn.Query("SELECT artist, album, title, duration, path FROM musicas")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var musicas []MusicFile
	for rows.Next() {
		var m MusicFile
		var durationSeconds int64
		err := rows.Scan(&m.Artist, &m.Album, &m.Title, &durationSeconds, &m.Path)
		if err != nil {
			return nil, err
		}
		m.Duration = time.Duration(durationSeconds) * time.Second
		musicas = append(musicas, m)
	}
	return musicas, nil
}

// GetArtists returns a list of distinct artists
func (db *DB) GetArtists() ([]string, error) {
	rows, err := db.conn.Query("SELECT DISTINCT artist FROM musicas WHERE artist IS NOT NULL AND artist != ''")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []string
	for rows.Next() {
		var artist string
		err := rows.Scan(&artist)
		if err != nil {
			return nil, err
		}
		artists = append(artists, artist)
	}
	return artists, nil
}

// GetAlbumsByArtist returns a list of albums for a given artist
func (db *DB) GetAlbumsByArtist(artist string) ([]string, error) {
	rows, err := db.conn.Query("SELECT DISTINCT album FROM musicas WHERE artist = ? AND album IS NOT NULL AND album != ''", artist)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []string
	for rows.Next() {
		var album string
		err := rows.Scan(&album)
		if err != nil {
			return nil, err
		}
		albums = append(albums, album)
	}
	return albums, nil
}

// SearchMusicByGenre returns music files matching the genre
func (db *DB) SearchMusicByGenre(genre string) ([]MusicFile, error) {
	rows, err := db.conn.Query("SELECT artist, album, title, duration, path FROM musicas WHERE genre LIKE ?", "%"+genre+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var musicas []MusicFile
	for rows.Next() {
		var m MusicFile
		var durationSeconds int64
		err := rows.Scan(&m.Artist, &m.Album, &m.Title, &durationSeconds, &m.Path)
		if err != nil {
			return nil, err
		}
		m.Duration = time.Duration(durationSeconds) * time.Second
		musicas = append(musicas, m)
	}
	return musicas, nil
}

// GetSongsByArtistAndAlbum returns songs for a given artist and album
func (db *DB) GetSongsByArtistAndAlbum(artist, album string) ([]MusicFile, error) {
	rows, err := db.conn.Query("SELECT artist, album, title, duration, path FROM musicas WHERE artist = ? AND album = ?", artist, album)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []MusicFile
	for rows.Next() {
		var m MusicFile
		var durationSeconds int64
		err := rows.Scan(&m.Artist, &m.Album, &m.Title, &durationSeconds, &m.Path)
		if err != nil {
			return nil, err
		}
		m.Duration = time.Duration(durationSeconds) * time.Second
		songs = append(songs, m)
	}
	return songs, nil
}

// Verificar usuario y contraseña
func (db *DB) VerifyUser(email, clave string) (bool, error) {
	var hashed string
	err := db.conn.QueryRow("SELECT clave FROM usuarios WHERE email=?", email).Scan(&hashed)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(clave))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Activar usuario
func (db *DB) ActivateUser(email string) error {
	_, err := db.conn.Exec("UPDATE usuarios SET activo=TRUE WHERE email=?", email)
	return err
}

// Obtener lista de usuarios
func (db *DB) GetUsers() ([]Usuario, error) {
	rows, err := db.conn.Query("SELECT id, nombre, email, activo FROM usuarios")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []Usuario
	for rows.Next() {
		var u Usuario
		err := rows.Scan(&u.ID, &u.Nombre, &u.Email, &u.Activo)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, u)
	}
	return usuarios, nil
}
