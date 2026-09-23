package internal

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository define la interfaz del repositorio de catálogo (alumnos, armas, ubicaciones)
type Repository interface {
	// GetStudent devuelve el alumno con ese id
	GetStudent(ctx context.Context, id string) (Student, error)

	// ListStudents devuelve todos los alumnos ordenados por nombre
	ListStudents(ctx context.Context) ([]Student, error)

	// GetWeapon devuelve el arma con ese id
	GetWeapon(ctx context.Context, id string) (Weapon, error)

	// ListWeapons devuelve todas las armas ordenadas por nombre
	ListWeapons(ctx context.Context) ([]Weapon, error)

	// GetLocation devuelve la ubicación con ese id
	GetLocation(ctx context.Context, id string) (Location, error)

	// ListLocations devuelve todas las ubicaciones ordenadas por nombre
	ListLocations(ctx context.Context) ([]Location, error)
}

// CatalogRepository implementa Repository sobre las tablas de Postgres
type CatalogRepository struct {
	pool *pgxpool.Pool
}

// Comprueba que CatalogRepository cumple la interfaz Repository
var _ Repository = (*CatalogRepository)(nil)

// NewCatalogRepository crea el repositorio sobre un pool abierto de postgres
func NewCatalogRepository(pool *pgxpool.Pool) *CatalogRepository {
	return &CatalogRepository{pool: pool}
}

//* Implementación de la interfaz

// GetStudent devuelve un alumno filtrado por ID
//
// Param - ctx: Contexto para la query
// Param - id: UUID del alumno ya validado por el servicio
// Returns - Student/error: El alumno recuperado de la base de datos o el error si lo hay
func (cr *CatalogRepository) GetStudent(ctx context.Context, id string) (Student, error) {
	// Crea una variable de tipo Student
	var s Student

	// Busca al alumno en la tabla students y escanea el resultado en la variable
	err := cr.pool.QueryRow(ctx,
		`SELECT id, name, strength, agility, stamina FROM students WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.Strength, &s.Agility, &s.Stamina)

	// Si ese alumno no se encuentra en la tabla devuelve un error not found
	if errors.Is(err, pgx.ErrNoRows) {
		return Student{}, fmt.Errorf("student %s: %w", id, ErrNotFound)
	}

	// Devuelve cualquier otro fallo es de la base de datos distinto a not found
	if err != nil {
		return Student{}, fmt.Errorf("querying student %s: %w", id, err)
	}

	// Si todo va bien devuelve al alumno recuperado de la base de datos
	return s, nil
}

// ListStudents devuelve todos los alumnos ordenados por nombre
//
// Param - ctx: Contexto para la query
// Returns - []Student/error: Los alumnos recuperados de la base de datos o el error si lo hay
func (cr *CatalogRepository) ListStudents(ctx context.Context) ([]Student, error) {
	// Recoge a todos los alumnos ordenados por nombre
	rows, err := cr.pool.Query(ctx,
		`SELECT id, name, strength, agility, stamina FROM students ORDER BY name`,
	)

	if err != nil {
		return nil, fmt.Errorf("querying students: %w", err)
	}

	// Difiere el cierre de las filas
	defer rows.Close()

	// Crea un slice de Students
	students := []Student{}

	// Escanea cada fila de la tabla en el slice recién creado
	for rows.Next() {
		var s Student

		if err := rows.Scan(&s.ID, &s.Name, &s.Strength, &s.Agility, &s.Stamina); err != nil {
			return nil, fmt.Errorf("scanning student: %w", err)
		}

		students = append(students, s)
	}

	// Si hay un error de lectura de filas devuelve el error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating students: %w", err)
	}

	// Si todo va bien devuelve el slice de students
	return students, nil
}

// GetWeapon devuelve el arma filtrado por ID
//
// Param - ctx: Contexto para la query
// Param - id: UUID del arma ya validado por el servicio
// Returns - Weapon/error: El arma recuperada de la base de datos o el error si lo hay
func (cr *CatalogRepository) GetWeapon(ctx context.Context, id string) (Weapon, error) {
	// Crea una variable de tipo Student
	var w Weapon

	// Busca al arma en la tabla students y escanea el resultado en la variable
	err := cr.pool.QueryRow(ctx,
		`SELECT id, name, damage, accuracy FROM weapons WHERE id = $1`, id,
	).Scan(&w.ID, &w.Name, &w.Damage, &w.Accuracy)

	// Si ese arma no se encuentra en la tabla devuelve un error not found
	if errors.Is(err, pgx.ErrNoRows) {
		return Weapon{}, fmt.Errorf("weapon %s: %w", id, ErrNotFound)
	}

	// Cualquier otro fallo es de la base de datos
	if err != nil {
		return Weapon{}, fmt.Errorf("querying weapon %s: %w", id, err)
	}

	// Si todo va bien devuelve el arma recuperada de la base de datos
	return w, nil
}

// ListStudents devuelve todas los armas ordenados por nombre
//
// Param - ctx: Contexto para la query
// Returns - []Weapon/error: Las armas recuperadas de la base de datos o el error si lo hay
func (cr *CatalogRepository) ListWeapons(ctx context.Context) ([]Weapon, error) {
	// Recoge todas las armas ordenadas por nombre
	rows, err := cr.pool.Query(ctx,
		`SELECT id, name, damage, accuracy FROM weapons ORDER BY name`,
	)

	if err != nil {
		return nil, fmt.Errorf("querying weapons: %w", err)
	}

	// Difiere el cierre de las filas
	defer rows.Close()

	// Crea un slice de Weapons
	weapons := []Weapon{}

	// Escanea cada fila de la tabla en el slice recién creado
	for rows.Next() {
		var w Weapon

		if err := rows.Scan(&w.ID, &w.Name, &w.Damage, &w.Accuracy); err != nil {
			return nil, fmt.Errorf("scanning weapon: %w", err)
		}

		weapons = append(weapons, w)
	}

	// Si hay un error de lectura de filas devuelve el error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating weapons: %w", err)
	}

	// Si todo va bien devuelve el slice de weapons
	return weapons, nil
}

// GetLocation devuelve la ubicación filtrada por ID
//
// Param - ctx: Contexto para la query
// Param - id: UUID de la ubicación ya validado por el servicio
// Returns - Location/error: La ubicación recuperada de la base de datos o el error si lo hay
func (cr *CatalogRepository) GetLocation(ctx context.Context, id string) (Location, error) {
	// Crea una variable de tipo Location
	var l Location

	// Busca al arma en la tabla students y escanea el resultado en la variable
	err := cr.pool.QueryRow(ctx,
		`SELECT id, name, description FROM locations WHERE id = $1`, id,
	).Scan(&l.ID, &l.Name, &l.Description)

	// Si ese arma no se encuentra en la tabla devuelve un error not found
	if errors.Is(err, pgx.ErrNoRows) {
		return Location{}, fmt.Errorf("location %s: %w", id, ErrNotFound)
	}

	// Cualquier otro fallo es de la base de datos
	if err != nil {
		return Location{}, fmt.Errorf("querying location %s: %w", id, err)
	}

	// Si todo va bien devuelve el arma recuperada de la base de datos
	return l, nil
}

// ListLocations devuelve todas los ubicaciones ordenados por nombre
//
// Param - ctx: Contexto para la query
// Returns - []Location/error: Las ubicaciones recuperadas de la base de datos o el error si lo hay
func (cr *CatalogRepository) ListLocations(ctx context.Context) ([]Location, error) {
	// Recoge todas las ubicaciones ordenadas por nombre
	rows, err := cr.pool.Query(ctx,
		`SELECT id, name, description FROM locations ORDER BY name`,
	)

	if err != nil {
		return nil, fmt.Errorf("querying locations: %w", err)
	}

	// Difiere el cierre de las filas
	defer rows.Close()

	// Crea un slice de Location
	locations := []Location{}

	// Escanea cada fila de la tabla en el slice recién creado
	for rows.Next() {
		var l Location

		if err := rows.Scan(&l.ID, &l.Name, &l.Description); err != nil {
			return nil, fmt.Errorf("scanning location: %w", err)
		}

		locations = append(locations, l)
	}

	// Si hay un error de lectura de filas devuelve el error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating locations: %w", err)
	}

	// Si todo va bien devuelve el slice de locations
	return locations, nil
}
