package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var db *DB
var tpl *template.Template

// Estructuras de datos para la plantilla
type User struct {
	ID     int
	Nombre string
	Email  string
}

type Song struct {
	ID     string
	Title  string
	Artist string
	Album  string
}

type Album struct {
	Name  string
	Songs []Song
}

type Artist struct {
	Name   string
	Albums []Album
}

// Estructura para pasar mensajes a las plantillas
type PageData struct {
	Success bool
	Message string
}

var funcMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}

func main() {
	var err error
	// Conexión a la base de datos
	db, err = NewDB("root", "03042008", "localhost", "hify_player")
	if err != nil {
		log.Fatalf("Error fatal conectando a la base de datos: %v", err)
	}
	defer db.Close()
	// Inicializar tablas si no existen
	if err := db.Init(); err != nil {
		log.Fatalf("Error al inicializar la base de datos: %v", err)
	}

	tpl = template.New("").Funcs(funcMap)
	tpl, err = tpl.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error fatal cargando las plantillas HTML: %v", err)
	}

	// Rutas
	http.HandleFunc("/", loginHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler) // NUEVA RUTA
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/api/song/", songStreamHandler)

	fmt.Println("Servidor iniciado en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ---- Controladores ----

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("session_token"); err == nil {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	// Recoger mensaje de la URL si existe (desde el registro)
	msg := r.URL.Query().Get("msg")
	data := PageData{
		Success: true,
		Message: msg,
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		verified, err := db.VerifyUser(email, password)
		if err != nil {
			log.Printf("Error al verificar usuario: %v", err)
			http.Error(w, "Error del servidor", http.StatusInternalServerError)
			return
		}

		if !verified {
			tpl.ExecuteTemplate(w, "login.html", PageData{Success: false, Message: "Usuario o clave incorrectos"})
			return
		}

		// Si la verificación es exitosa, obtener ID del usuario para la cookie
		var userID int
		// NOTA: Asumimos que el email es único.
		err = db.conn.QueryRow("SELECT id FROM usuarios WHERE email = ?", email).Scan(&userID)
		if err != nil {
			log.Printf("Error al obtener ID de usuario tras verificación: %v", err)
			http.Error(w, "Error del servidor", http.StatusInternalServerError)
			return
		}

		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{Name: "session_token", Value: fmt.Sprintf("%d", userID), Expires: expiration, Path: "/"}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.html", data)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		nombre := r.FormValue("nombre")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if nombre == "" || email == "" || password == "" {
			tpl.ExecuteTemplate(w, "register.html", "Todos los campos son obligatorios.")
			return
		}

		err := db.AddUser(nombre, email, password)
		if err != nil {
			// Comprobar si el error es porque el usuario ya existe
			if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
				tpl.ExecuteTemplate(w, "register.html", "El correo electrónico ya está registrado.")
			} else {
				log.Printf("Error al añadir usuario: %v", err)
				tpl.ExecuteTemplate(w, "register.html", "Ocurrió un error al crear la cuenta.")
			}
			return
		}

		// Redirigir a login con mensaje de éxito
		http.Redirect(w, r, "/login?msg=¡Cuenta creada con éxito! Por favor, inicia sesión.", http.StatusSeeOther)
		return
	}

	// Si es GET, solo mostrar la página
	tpl.ExecuteTemplate(w, "register.html", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Validar que el ID de la cookie es un número
	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		// Cookie inválida, destruir y redirigir
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Expires: time.Now().Add(-1 * time.Hour), Path: "/"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var user User
	err = db.conn.QueryRow("SELECT id, nombre, email FROM usuarios WHERE id = ?", userID).Scan(&user.ID, &user.Nombre, &user.Email)
	if err != nil {
		log.Printf("Error al buscar datos del usuario (ID: %d): %v", userID, err)
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Expires: time.Now().Add(-1 * time.Hour), Path: "/"})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	library, err := getMusicLibrary()
	if err != nil {
		log.Printf("Error al obtener la biblioteca de música: %v", err)
		http.Error(w, "Error obteniendo la biblioteca de música", http.StatusInternalServerError)
		return
	}

	data := struct {
		User    User
		Library []Artist
	}{
		User:    user,
		Library: library,
	}
	tpl.ExecuteTemplate(w, "home.html", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Expires: time.Now().Add(-1 * time.Hour), Path: "/"})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func songStreamHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("session_token"); err != nil {
		http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
		return
	}

	songIDStr := strings.TrimPrefix(r.URL.Path, "/api/song/")
	songID, err := strconv.Atoi(songIDStr)
	if err != nil {
		log.Printf("ID de canción inválido recibido: %s", songIDStr)
		http.Error(w, "ID de canción inválido", http.StatusBadRequest)
		return
	}

	var filePath string
	err = db.conn.QueryRow("SELECT path FROM musicas WHERE id = ?", songID).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Printf("Error al buscar la ruta de la canción (ID: %d): %v", songID, err)
			http.Error(w, "Error del servidor", http.StatusInternalServerError)
		}
		return
	}
	http.ServeFile(w, r, filePath)
}

func getMusicLibrary() ([]Artist, error) {
	query := "SELECT id, title, artist, album FROM musicas WHERE artist IS NOT NULL AND artist != '' ORDER BY artist, album, title"
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	artistsMap := make(map[string]map[string][]Song)

	for rows.Next() {
		var songID, title, artistName string
		var albumName sql.NullString

		if err := rows.Scan(&songID, &title, &artistName, &albumName); err != nil {
			log.Printf("Error al escanear fila de música: %v", err)
			continue
		}

		var currentAlbumName string
		if albumName.Valid && albumName.String != "" {
			currentAlbumName = albumName.String
		} else {
			currentAlbumName = "Varios"
		}

		if _, ok := artistsMap[artistName]; !ok {
			artistsMap[artistName] = make(map[string][]Song)
		}

		song := Song{
			ID:     songID,
			Title:  title,
			Artist: artistName,
			Album:  currentAlbumName,
		}
		artistsMap[artistName][currentAlbumName] = append(artistsMap[artistName][currentAlbumName], song)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	var library []Artist
	for artistName, albumsMap := range artistsMap {
		var albums []Album
		for albumName, songs := range albumsMap {
			albums = append(albums, Album{Name: albumName, Songs: songs})
		}
		library = append(library, Artist{Name: artistName, Albums: albums})
	}
	return library, nil
}
