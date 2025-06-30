package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	currentPlaylist  []MusicFile
	currentSongIndex int = -1
)

func main() {
	// Conectar a la base de datos
	db, err := NewDB("root", "03042008", "localhost:3306", "hify_player")
	if err != nil {
		fmt.Println("Error conectando a la base de datos:", err)
		return
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		fmt.Println("Error inicializando la base de datos:", err)
		return
	}

	// Escanear directorio de música y guardar metadatos en la base de datos
	musicas, err := ScanMusicDirectory("C:\\Users\\Usuario\\Documents\\GitHub\\PROYECTO-MUSICA-HIRES\\PROYECTO-MUSICA-HIRES\\Musicas")
	if err != nil {
		fmt.Println("Error al escanear música:", err)
	} else {
		for _, m := range musicas {
			err := db.AddOrUpdateMusic(m)
			if err != nil {
				fmt.Printf("Error al guardar metadatos de %s: %v\n", m.Path, err)
			}
		}
		fmt.Println("Escaneo de música completado.")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n===================================")
		fmt.Println("      Bienvenido a hifiplayer      ")
		fmt.Println("===================================")
		fmt.Println("1. Ingresar")
		fmt.Println("2. Registrarse")
		fmt.Println("3. Salir")
		fmt.Print("Seleccione una opción: ")

		opcionStr, _ := reader.ReadString('\n')
		opcionStr = strings.TrimSpace(opcionStr)

		switch opcionStr {
		case "1":
			if ingresar(db, reader) {
				fmt.Println("Acceso concedido")
				mostrarUsuarios(db)
				menuUsuario(db, reader)
			} else {
				fmt.Println("Usuario o clave incorrectos")
			}
		case "2":
			registrar(db, reader)
		case "3":
			fmt.Println("Gracias por usar hifiplayer. ¡Adiós!")
			os.Exit(0)
		default:
			fmt.Println("Opción inválida. Intente de nuevo.")
		}
	}
}

func ingresar(db *DB, reader *bufio.Reader) bool {
	fmt.Print("Ingrese su email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Ingrese su clave: ")
	clave, _ := reader.ReadString('\n')
	clave = strings.TrimSpace(clave)

	ok, err := db.VerifyUser(email, clave)
	if err != nil {
		fmt.Println("Error verificando usuario:", err)
		return false
	}
	if !ok {
		return false
	}

	err = db.ActivateUser(email)
	if err != nil {
		fmt.Println("Error activando usuario:", err)
		return false
	}

	return true
}

func registrar(db *DB, reader *bufio.Reader) {
	fmt.Print("Ingrese su nombre: ")
	nombre, _ := reader.ReadString('\n')
	nombre = strings.TrimSpace(nombre)

	fmt.Print("Ingrese su email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Ingrese su clave: ")
	clave, _ := reader.ReadString('\n')
	clave = strings.TrimSpace(clave)

	err := db.AddUser(nombre, email, clave)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Usuario registrado con éxito!")
	}
}

func mostrarUsuarios(db *DB) {
	usuarios, err := db.GetUsers()
	if err != nil {
		fmt.Println("Error obteniendo usuarios:", err)
		return
	

		
	}
	fmt.Println("--------------------------\n")
}

func menuUsuario(db *DB, reader *bufio.Reader) {
	for {
		fmt.Println("\n===================================")
		fmt.Println("         Menú Principal           ")
		fmt.Println("===================================")
		fmt.Println("1. Explorar Artistas y Álbumes")
		fmt.Println("2. Cerrar sesión")
		fmt.Print("Seleccione una opción: ")

		opcionStr, _ := reader.ReadString('\n')
		opcionStr = strings.TrimSpace(opcionStr)

		switch opcionStr {
		case "1":
			explorarMusica(db, reader)
		case "2":
			fmt.Println("Cerrando sesión...")
			StopMusic() // Detener la música al cerrar sesión
			return
		default:
			fmt.Println("Opción inválida. Intente de nuevo.")
		}
	}
}

func explorarMusica(db *DB, reader *bufio.Reader) {
	// Paso 1: Listar y seleccionar artista
	artistas, err := db.GetArtists()
	if err != nil {
		fmt.Println("Error obteniendo artistas:", err)
		return
	}
	if len(artistas) == 0 {
		fmt.Println("No se encontraron artistas en la base de datos.")
		return
	}

	fmt.Println("\n--- Artistas Disponibles ---")
	for i, artista := range artistas {
		fmt.Printf("%d. %s\n", i+1, artista)
	}
	fmt.Print("Seleccione un artista por número (o '0' para volver): ")
	artistaNumStr, _ := reader.ReadString('\n')
	artistaNum, err := strconv.Atoi(strings.TrimSpace(artistaNumStr))
	if err != nil || artistaNum < 0 || artistaNum > len(artistas) {
		fmt.Println("Selección inválida.")
		return
	}
	if artistaNum == 0 {
		return
	}
	artistaSeleccionado := artistas[artistaNum-1]

	// Paso 2: Listar y seleccionar álbum
	albumes, err := db.GetAlbumsByArtist(artistaSeleccionado)
	if err != nil {
		fmt.Println("Error obteniendo álbumes:", err)
		return
	}
	if len(albumes) == 0 {
		fmt.Printf("No se encontraron álbumes para el artista '%s'.\n", artistaSeleccionado)
		return
	}

	fmt.Printf("\n--- Álbumes de %s ---\n", artistaSeleccionado)
	for i, album := range albumes {
		fmt.Printf("%d. %s\n", i+1, album)
	}
	fmt.Print("Seleccione un álbum por número (o '0' para volver): ")
	albumNumStr, _ := reader.ReadString('\n')
	albumNum, err := strconv.Atoi(strings.TrimSpace(albumNumStr))
	if err != nil || albumNum < 0 || albumNum > len(albumes) {
		fmt.Println("Selección inválida.")
		return
	}
	if albumNum == 0 {
		return
	}
	albumSeleccionado := albumes[albumNum-1]

	// Paso 3: Listar y seleccionar canción
	canciones, err := db.GetSongsByArtistAndAlbum(artistaSeleccionado, albumSeleccionado)
	if err != nil {
		fmt.Println("Error obteniendo canciones:", err)
		return
	}
	if len(canciones) == 0 {
		fmt.Println("No se encontraron canciones en este álbum.")
		return
	}
	currentPlaylist = canciones // Actualizar la lista de reproducción actual

	fmt.Printf("\n--- Canciones de '%s' - '%s' ---\n", artistaSeleccionado, albumSeleccionado)
	for i, cancion := range canciones {
		fmt.Printf("%d. %s\n", i+1, cancion.Title)
	}
	fmt.Print("Seleccione una canción para reproducir (o '0' para volver): ")
	cancionNumStr, _ := reader.ReadString('\n')
	cancionNum, err := strconv.Atoi(strings.TrimSpace(cancionNumStr))
	if err != nil || cancionNum < 0 || cancionNum > len(canciones) {
		fmt.Println("Selección inválida.")
		return
	}
	if cancionNum == 0 {
		return
	}

	currentSongIndex = cancionNum - 1
	err = PlayMusicFile(currentPlaylist[currentSongIndex].Path)
	if err != nil {
		fmt.Println("Error al reproducir canción:", err)
	} else {
		playerMenu(reader)
	}
}

func playerMenu(reader *bufio.Reader) {
	for {
		fmt.Println("\n===================================")
		fmt.Println("         Menú de Reproducción      ")
		fmt.Println("===================================")
		fmt.Println("1. Retroceder (10s)")
		fmt.Println("2. Play/Pausar")
		fmt.Println("3. Adelantar (10s)")
		fmt.Println("4. Siguiente canción")
		fmt.Println("5. Canción anterior")
		fmt.Println("6. Ajustar volumen")
		fmt.Println("7. Detener y volver al menú")
		fmt.Print("Seleccione una opción: ")

		opcionStr, _ := reader.ReadString('\n')
		opcionStr = strings.TrimSpace(opcionStr)

		switch opcionStr {
		case "1":
			SeekMusic(-10 * time.Second)
		case "2":
			TogglePlayPause()
		case "3":
			SeekMusic(10 * time.Second)
		case "4":
			if currentSongIndex != -1 && currentSongIndex < len(currentPlaylist)-1 {
				currentSongIndex++
				err := PlayMusicFile(currentPlaylist[currentSongIndex].Path)
				if err != nil {
					fmt.Println("Error al reproducir la siguiente canción:", err)
				}
			} else {
				fmt.Println("No hay siguiente canción en la lista.")
			}
		case "5":
			if currentSongIndex > 0 {
				currentSongIndex--
				err := PlayMusicFile(currentPlaylist[currentSongIndex].Path)
				if err != nil {
					fmt.Println("Error al reproducir la canción anterior:", err)
				}
			} else {
				fmt.Println("No hay canción anterior en la lista.")
			}
		case "6":
			fmt.Print("Ingrese el volumen (0.0 a 1.0): ")
			volStr, _ := reader.ReadString('\n')
			volStr = strings.TrimSpace(volStr)
			vol, err := strconv.ParseFloat(volStr, 64)
			if err != nil || vol < 0.0 || vol > 1.0 {
				fmt.Println("Volumen inválido. Ingrese un valor entre 0.0 y 1.0.")
			} else {
				SetVolume(vol)
			}
		case "7":
			StopMusic()
			currentSongIndex = -1 // Reiniciar índice para indicar que no hay canción en reproducción
			return                // Salir del menú de reproducción
		default:
			fmt.Println("Opción inválida. Intente de nuevo.")
		}
	}
}
