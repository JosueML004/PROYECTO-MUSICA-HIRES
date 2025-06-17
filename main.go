package main

import (
	"fmt"
	"os"
)

// Estructura para representar un Usuario
type Usuario struct {
	nombre string
	email  string
	clave  string
	activo bool
}

// Método para activar el usuario
func (u *Usuario) Activar() {
	u.activo = true
}

// Método para obtener el nombre del usuario
func (u *Usuario) GetNombre() string {
	return u.nombre
}

// Estructura para representar Contenido
type Contenido struct {
	titulo      string
	descripcion string
	disponible  bool
}

// Método para marcar contenido como disponible
func (c *Contenido) MarcarComoDisponible() {
	c.disponible = true
}

// Método para marcar contenido como no disponible
func (c *Contenido) MarcarComoNoDisponible() {
	c.disponible = false
}

// Estructura para representar la Plataforma de Streaming
type Plataforma struct {
	usersPasswords *UsersPasswords
	contenido      map[string]*Contenido
}

// Constructor para Plataforma
func NewPlataforma() *Plataforma {
	return &Plataforma{
		usersPasswords: NewUsersPasswords(),
		contenido:      make(map[string]*Contenido),
	}
}

// Método para agregar contenido a la plataforma
func (p *Plataforma) AgregarContenido(titulo string, descripcion string) {
	p.contenido[titulo] = &Contenido{titulo: titulo, descripcion: descripcion, disponible: false}
}

func main() {
	plataforma := NewPlataforma()

	for {
		fmt.Println("===================================")
		fmt.Println("      Bienvenido a hifiplayer      ")
		fmt.Println("===================================")
		fmt.Println("1. Ingresar")
		fmt.Println("2. Registrarse")
		fmt.Println("3. Salir")
		fmt.Print("Seleccione una opción: ")

		var opcion int
		_, err := fmt.Scanln(&opcion)
		if err != nil {
			fmt.Println("Entrada inválida. Intente de nuevo.")
			continue
		}

		switch opcion {
		case 1:
			if ingresar(plataforma) {
				fmt.Println("Acceso concedido")
				mostrarUsuarios(plataforma)
				// Aquí puedes añadir el menú o funcionalidades para usuarios autenticados
			} else {
				fmt.Println("Usuario o clave incorrectos")
			}
		case 2:
			registrar(plataforma)
		case 3:
			fmt.Println("Gracias por usar hifiplayer. ¡Adiós!")
			os.Exit(0)
		default:
			fmt.Println("Opción inválida. Intente de nuevo.")
		}
	}
}

func ingresar(p *Plataforma) bool {
	fmt.Print("Ingrese su email: ")
	var email string
	_, err := fmt.Scanln(&email)
	if err != nil {
		fmt.Println("Error leyendo email.")
		return false
	}
	fmt.Print("Ingrese su clave: ")
	var clave string
	_, err = fmt.Scanln(&clave)
	if err != nil {
		fmt.Println("Error leyendo clave.")
		return false
	}

	if !p.usersPasswords.VerifyUser(email, clave) {
		return false
	}

	err = p.usersPasswords.ActivateUser(email)
	if err != nil {
		fmt.Println("Error activando usuario:", err)
		return false
	}

	return true
}

func registrar(p *Plataforma) {
	fmt.Print("Ingrese su nombre: ")
	var nombre string
	_, err := fmt.Scanln(&nombre)
	if err != nil {
		fmt.Println("Error leyendo nombre.")
		return
	}
	fmt.Print("Ingrese su email: ")
	var email string
	_, err = fmt.Scanln(&email)
	if err != nil {
		fmt.Println("Error leyendo email.")
		return
	}
	fmt.Print("Ingrese su clave: ")
	var clave string
	_, err = fmt.Scanln(&clave)
	if err != nil {
		fmt.Println("Error leyendo clave.")
		return
	}

	err = p.usersPasswords.AddUser(email, nombre, clave)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Usuario registrado con éxito!")
	}
}

func mostrarUsuarios(p *Plataforma) {
	usuarios := p.usersPasswords.ShowUsers()
	fmt.Println("\nUsuarios registrados:")
	for _, usuario := range usuarios {
		fmt.Printf("Nombre: %s, Email: %s, Activo: %t\n", usuario.GetNombre(), usuario.email, usuario.activo)
	}
	fmt.Println()
}
