package main

import (
	"errors"
	"strings"
	"sync"
)

// Estructura que guarda los usuarios y contraseñas registrados
type UsersPasswords struct {
	mu       sync.RWMutex
	usuarios map[string]*Usuario
}

// Constructor para UsersPasswords
func NewUsersPasswords() *UsersPasswords {
	return &UsersPasswords{
		usuarios: make(map[string]*Usuario),
	}
}

// Método para agregar un usuario con contraseña
func (up *UsersPasswords) AddUser(email, nombre, clave string) error {
	up.mu.Lock()
	defer up.mu.Unlock()

	emailLower := strings.ToLower(email)
	if _, existe := up.usuarios[emailLower]; existe {
		return errors.New("el usuario ya existe")
	}
	up.usuarios[emailLower] = &Usuario{
		nombre: nombre,
		email:  emailLower,
		clave:  clave,
		activo: false,
	}
	return nil
}

// Método para obtener un usuario
func (up *UsersPasswords) GetUser(email string) (*Usuario, bool) {
	up.mu.RLock()
	defer up.mu.RUnlock()

	usuario, existe := up.usuarios[strings.ToLower(email)]
	return usuario, existe
}

// Método para verificar usuario y contraseña
func (up *UsersPasswords) VerifyUser(email, clave string) bool {
	up.mu.RLock()
	defer up.mu.RUnlock()

	usuario, existe := up.usuarios[strings.ToLower(email)]
	if !existe {
		return false
	}
	if usuario.clave != clave {
		return false
	}
	return true
}

// Método para activar un usuario
func (up *UsersPasswords) ActivateUser(email string) error {
	up.mu.Lock()
	defer up.mu.Unlock()

	usuario, existe := up.usuarios[strings.ToLower(email)]
	if !existe {
		return errors.New("usuario no encontrado")
	}
	usuario.Activar()
	return nil
}

// Método para mostrar todos los usuarios
func (up *UsersPasswords) ShowUsers() []Usuario {
	up.mu.RLock()
	defer up.mu.RUnlock()

	var usuariosList []Usuario
	for _, usuario := range up.usuarios {
		usuariosList = append(usuariosList, *usuario)
	}
	return usuariosList
}
