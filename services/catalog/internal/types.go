package internal

import "errors"

// Errores custom
//
// ErrNotFound indica que no existe ninguna fila con el id pedido
// ErrInvalidID indica que el id recibido no es un UUID válido
var (
	ErrNotFound  = errors.New("not found")
	ErrInvalidID = errors.New("invalid id")
)

// Student representa un alumno participante
type Student struct {
	ID       string
	Name     string
	Strength int32
	Agility  int32
	Stamina  int32
}

// Weapon representa un arma
type Weapon struct {
	ID       string
	Name     string
	Damage   int32
	Accuracy int32
}

// Location representa una ubicación donde puede desarrollarse la batalla
type Location struct {
	ID          string
	Name        string
	Description string
}
