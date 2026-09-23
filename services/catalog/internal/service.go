package internal

import (
	"context"

	"github.com/google/uuid"
)

// Service define la interfaz del servicio de catálogo (alumnos, armas, ubicaciones)
type Service interface {
	// GetStudent devuelve el alumno con ese id
	GetStudent(ctx context.Context, id string) (Student, error)

	// ListStudents devuelve todos los alumnos ordenados por nombre
	ListStudents(ctx context.Context) ([]Student, error)

	// GetWeapon devuelve el arma con ese id
	GetWeapon(ctx context.Context, id string) (Weapon, error)

	// ListWeapons devuelve todas las armas ordenadas por nombre
	ListWeapons(ctx context.Context) ([]Weapon, error)

	// GetLocation devuelve la localización con ese id
	GetLocation(ctx context.Context, id string) (Location, error)

	// ListLocations devuelve todas las localizaciones ordenadas por nombre
	ListLocations(ctx context.Context) ([]Location, error)
}

// CatalogService implementa Service validando la entrada antes de llegar al repositorio
type CatalogService struct {
	repo Repository
}

// Comprueba que CatalogService cumple la interfaz Service
var _ Service = (*CatalogService)(nil)

// NewCatalogService crea el servicio sobre el repositorio sobre el repositorio recibido por inyección
func NewCatalogService(repo Repository) *CatalogService {
	return &CatalogService{repo: repo}
}

//* Implementación de la interfaz

// GetStudent devuelve un alumno filtrado por ID
//
// Param - ctx: Contexto enviado al repositorio
// Param - id: UUID del alumno
// Returns - Student/error: El alumno recuperado de la base de datos o el error si lo hay
func (s *CatalogService) GetStudent(ctx context.Context, id string) (Student, error) {
	// Comprueba que el UUID tenga el formato correcto. Devuelve error si está mal
	parsed, err := uuid.Parse(id)
	if err != nil {
		return Student{}, ErrInvalidID
	}

	// Llama al repositorio para obtener el alumno filtrado o el error
	return s.repo.GetStudent(ctx, parsed.String())
}

// ListStudents devuelve todos los alumnos ordenados por nombre
//
// Param - ctx: Contexto enviado al repositorio
// Returns - []Student/error: Los alumnos recuperados de la base de datos o el error si lo hay
func (s *CatalogService) ListStudents(ctx context.Context) ([]Student, error) {
	// Llama al repositorio para obtener el listado de alumnos
	return s.repo.ListStudents(ctx)
}

// GetWeapon devuelve un arma filtrada por ID
//
// Param - ctx: Contexto enviado al repositorio
// Param - id: UUID del arma
// Returns - Weapon/error: El arma recuperado de la base de datos o el error si lo hay
func (s *CatalogService) GetWeapon(ctx context.Context, id string) (Weapon, error) {
	// Comprueba que el UUID tenga el formato correcto. Devuelve error si está mal
	parsed, err := uuid.Parse(id)
	if err != nil {
		return Weapon{}, ErrInvalidID
	}

	// Llama al repositorio para obtener el arma filtrada o el error
	return s.repo.GetWeapon(ctx, parsed.String())
}

// ListWeapons devuelve todas los armas ordenadas por nombre
//
// Param - ctx: Contexto enviado al repositorio
// Returns - []Weapon/error: Las armas recuperadas de la base de datos o el error si lo hay
func (s *CatalogService) ListWeapons(ctx context.Context) ([]Weapon, error) {
	// Llama al repositorio para obtener el listado de armas
	return s.repo.ListWeapons(ctx)
}

// GetLocation devuelve una ubicación filtrada por ID
//
// Param - ctx: Contexto enviado al repositorio
// Param - id: UUID de la ubicación
// Returns - Location/error: La ubicación recuperado de la base de datos o el error si lo hay
func (s *CatalogService) GetLocation(ctx context.Context, id string) (Location, error) {
	// Comprueba que el UUID tenga el formato correcto. Devuelve error si está mal
	parsed, err := uuid.Parse(id)
	if err != nil {
		return Location{}, ErrInvalidID
	}

	// Llama al repositorio para obtener la ubicación filtrada o el error
	return s.repo.GetLocation(ctx, parsed.String())
}

// ListLocations devuelve todas los ubicaciones ordenadas por nombre
//
// Param - ctx: Contexto enviado al repositorio
// Param - id: UUID de la ubicación
// Returns - []Location/error: Las ubicaciones recuperadas de la base de datos o el error si lo hay
func (s *CatalogService) ListLocations(ctx context.Context) ([]Location, error) {
	// Llama al repositorio para obtener el listado de ubicaciones
	return s.repo.ListLocations(ctx)
}
