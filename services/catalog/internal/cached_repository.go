package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

// redisTimeout limita cada operación contra Redis para que una caché lenta no frene la RPC
const redisTimeout = 200 * time.Millisecond

// Claves de caché de los listados completos
const (
	studentsKey  = "catalog:students:all"
	weaponsKey   = "catalog:weapons:all"
	locationsKey = "catalog:locations:all"
)

// CachedRepository implementa Repository con cache-aside en Redis
type CachedRepository struct {
	next  Repository
	rdb   *redis.Client
	ttl   time.Duration
	group singleflight.Group
	log   *slog.Logger
}

// Comprueba que CachedRepository cumple la interfaz Repository
var _ Repository = (*CachedRepository)(nil)

// NewCachedRepository crea el repositorio con caché sobre el repositorio recibido por inyección
func NewCachedRepository(next Repository, rdb *redis.Client, ttl time.Duration, log *slog.Logger) *CachedRepository {
	return &CachedRepository{next: next, rdb: rdb, ttl: ttl, log: log}
}

//* Implementación de la interfaz

// GetStudent devuelve un alumno filtrado por ID, desde Redis si está cacheado
//
// Param - ctx: Contexto de la llamada
// Param - id: UUID del alumno ya validado por el servicio
// Returns - Student/error: El alumno, o el error del repositorio en caso de que lo haya
func (cr *CachedRepository) GetStudent(ctx context.Context, id string) (Student, error) {
	// Crea el key de redis: tabla:Alumno:ID
	key := "catalog:student:" + id

	// Crea una variable de tipo Student
	var cachedStudent Student

	// Realiza la búsqueda en Redis y si lo encuentra devuelve el estudiante cacheado
	if cr.read(ctx, key, &cachedStudent) {
		return cachedStudent, nil
	}

	// Agrupa las peticiones concurrentes a la misma clave en una sola carga
	// Sin esto, cien peticiones concurrentes con la clave caducada harían cien consultas
	// idénticas a Postgres: Do deja pasar una sola y las demás esperan su resultado
	v, err, _ := cr.group.Do(key, func() (any, error) {
		// Obtiene el estudiante desde el repositoio
		student, err := cr.next.GetStudent(ctx, id)

		// Si se produce un error al obtener el resultado lo devuelve
		if err != nil {
			return nil, err
		}

		// Guarda el estudiante el Redis con la clave actualizada
		cr.write(ctx, key, student)

		return student, nil
	})

	// Si hay un error al hacer el do devuelve el error
	if err != nil {
		return Student{}, err
	}

	// Devuelve el estudiante recuperado de postgres
	return v.(Student), nil
}

// ListStudents devuelve todos los alumnos ordenados por nombre, desde Redis si están cacheados
//
// Param - ctx: Contexto de la llamada
// Returns - []Student/error: Los alumnos, o el error del repositorio en caso de que lo haya
func (cr *CachedRepository) ListStudents(ctx context.Context) ([]Student, error) {
	// Crea una variable de tipo slice de Student
	var cachedStudents []Student

	// Realiza la búsqueda en Redis y si lo encuentra devuelve el listado cacheado
	if cr.read(ctx, studentsKey, &cachedStudents) {
		return cachedStudents, nil
	}

	// Agrupa las peticiones concurrentes a la misma clave en una sola carga
	// Aquí importa más que en los Get: esta clave la comparten todos los clientes,
	// así que al caducar todas las peticiones en vuelo fallarían a la vez
	v, err, _ := cr.group.Do(studentsKey, func() (any, error) {
		// Obtiene los alumnos desde el repositorio
		students, err := cr.next.ListStudents(ctx)

		// Si se produce un error al obtener el resultado lo devuelve
		if err != nil {
			return nil, err
		}

		// Guarda el listado en Redis con la clave de los alumnos
		cr.write(ctx, studentsKey, students)

		return students, nil
	})

	// Si hay un error al hacer el Do devuelve el error
	if err != nil {
		return nil, err
	}

	// Devuelve el listado recuperado
	return v.([]Student), nil
}

// GetWeapon devuelve un arma filtrada por ID, desde Redis si está cacheada
//
// Param - ctx: Contexto de la llamada
// Param - id: UUID del arma ya validado por el servicio
// Returns - Weapon/error: El arma, o el error del repositorio en caso de que lo haya
func (cr *CachedRepository) GetWeapon(ctx context.Context, id string) (Weapon, error) {
	// Crea el key de redis: servicio:entidad:ID
	key := "catalog:weapon:" + id

	// Crea una variable de tipo Weapon
	var cachedWeapon Weapon

	// Realiza la búsqueda en Redis y si la encuentra devuelve el arma cacheada
	if cr.read(ctx, key, &cachedWeapon) {
		return cachedWeapon, nil
	}

	// Agrupa las peticiones concurrentes a la misma clave en una sola carga
	// Sin esto, cien peticiones concurrentes con la clave caducada harían cien consultas
	// idénticas a Postgres: Do deja pasar una sola y las demás esperan su resultado
	v, err, _ := cr.group.Do(key, func() (any, error) {
		// Obtiene el arma desde el repositorio
		weapon, err := cr.next.GetWeapon(ctx, id)

		// Si se produce un error al obtener el resultado lo devuelve
		if err != nil {
			return nil, err
		}

		// Guarda el arma en Redis con la clave construida
		cr.write(ctx, key, weapon)

		return weapon, nil
	})

	// Si hay un error al hacer el Do devuelve el error
	if err != nil {
		return Weapon{}, err
	}

	// Devuelve el arma recuperada
	return v.(Weapon), nil
}

// ListWeapons devuelve todas las armas ordenadas por nombre, desde Redis si están cacheadas
//
// Param - ctx: Contexto de la llamada
// Returns - []Weapon/error: Las armas, o el error del repositorio en caso de que lo haya
func (cr *CachedRepository) ListWeapons(ctx context.Context) ([]Weapon, error) {
	// Crea una variable de tipo slice de Weapon
	var cachedWeapons []Weapon

	// Realiza la búsqueda en Redis y si las encuentra devuelve el listado cacheado
	if cr.read(ctx, weaponsKey, &cachedWeapons) {
		return cachedWeapons, nil
	}

	// Agrupa las peticiones concurrentes a la misma clave en una sola carga
	// Aquí importa más que en los Get: esta clave la comparten todos los clientes,
	// así que al caducar todas las peticiones en vuelo fallarían a la vez
	v, err, _ := cr.group.Do(weaponsKey, func() (any, error) {
		// Obtiene las armas desde el repositorio
		weapons, err := cr.next.ListWeapons(ctx)

		// Si se produce un error al obtener el resultado lo devuelve
		if err != nil {
			return nil, err
		}

		// Guarda el listado en Redis con la clave de las armas
		cr.write(ctx, weaponsKey, weapons)

		return weapons, nil
	})

	// Si hay un error al hacer el Do devuelve el error
	if err != nil {
		return nil, err
	}

	// Devuelve el listado recuperado
	return v.([]Weapon), nil
}

// GetLocation devuelve una localización filtrada por ID, desde Redis si está cacheada
//
// Param - ctx: Contexto de la llamada
// Param - id: UUID de la localización ya validado por el servicio
// Returns - Location/error: La localización, o el error del repositorio en caso de que lo haya
func (cr *CachedRepository) GetLocation(ctx context.Context, id string) (Location, error) {
	// Crea el key de redis: servicio:entidad:ID
	key := "catalog:location:" + id

	// Crea una variable de tipo Location
	var cachedLocation Location

	// Realiza la búsqueda en Redis y si la encuentra devuelve la localización cacheada
	if cr.read(ctx, key, &cachedLocation) {
		return cachedLocation, nil
	}

	// Agrupa las peticiones concurrentes a la misma clave en una sola carga
	// Sin esto, cien peticiones concurrentes con la clave caducada harían cien consultas
	// idénticas a Postgres: Do deja pasar una sola y las demás esperan su resultado
	v, err, _ := cr.group.Do(key, func() (any, error) {
		// Obtiene la localización desde el repositorio
		location, err := cr.next.GetLocation(ctx, id)

		// Si se produce un error al obtener el resultado lo devuelve
		if err != nil {
			return nil, err
		}

		// Guarda la localización en Redis con la clave construida
		cr.write(ctx, key, location)

		return location, nil
	})

	// Si hay un error al hacer el Do devuelve el error
	if err != nil {
		return Location{}, err
	}

	// Devuelve la localización recuperada
	return v.(Location), nil
}

// ListLocations devuelve todas las localizaciones ordenadas por nombre, desde Redis si están cacheadas
//
// Param - ctx: Contexto de la llamada
// Returns - []Location/error: Las localizaciones, o el error del repositorio en caso de que lo haya
func (cr *CachedRepository) ListLocations(ctx context.Context) ([]Location, error) {
	// Crea una variable de tipo slice de Location
	var cachedLocations []Location

	// Realiza la búsqueda en Redis y si las encuentra devuelve el listado cacheado
	if cr.read(ctx, locationsKey, &cachedLocations) {
		return cachedLocations, nil
	}

	// Agrupa las peticiones concurrentes a la misma clave en una sola carga
	// Aquí importa más que en los Get: esta clave la comparten todos los clientes,
	// así que al caducar todas las peticiones en vuelo fallarían a la vez
	v, err, _ := cr.group.Do(locationsKey, func() (any, error) {
		// Obtiene las localizaciones desde el repositorio
		locations, err := cr.next.ListLocations(ctx)

		// Si se produce un error al obtener el resultado lo devuelve
		if err != nil {
			return nil, err
		}

		// Guarda el listado en Redis con la clave de las localizaciones
		cr.write(ctx, locationsKey, locations)

		return locations, nil
	})

	// Si hay un error al hacer el Do devuelve el error
	if err != nil {
		return nil, err
	}

	// Devuelve el listado recuperado
	return v.([]Location), nil
}

//* Funciones auxiliares de lectura y escritura en Redis

// read deserializa en dst (la variable creada, puntero) el valor de key y devuelve false si no está o si Redis falla
func (cr *CachedRepository) read(ctx context.Context, key string, dst any) bool {
	// Crea el contexto para redis y difiere el cierre
	rctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	// Lee los bytes guardados en la clave, tal cual se escribieron
	data, err := cr.rdb.Get(rctx, key).Bytes()

	switch {
	// La clave no está: no se loguea para no llenar el log
	case errors.Is(err, redis.Nil):
		return false

	// Redis falla o tarda: se avisa en el log sin propagar el error
	case err != nil:
		cr.log.Warn("cache read failed", "key", key, "error", err)
		return false
	}

	// Escanea dst con el valor cacheado; si el JSON está corrupto se avisa en el log y devuelve el error
	if err := json.Unmarshal(data, dst); err != nil {
		cr.log.Warn("cache decode failed", "key", key, "error", err)
		return false
	}

	// Si todo ha ido bien se devuelve true, el valor se ha recuperado de la caché
	return true
}

// write serializa v (valor) y lo guarda en redis bajo su key con el TTL, registrando cualquier fallo sin propagarlo
func (cr *CachedRepository) write(ctx context.Context, key string, v any) {
	// Serializa el valor a JSON
	data, err := json.Marshal(v)

	// Si hay un error se avisa en el log de redis y sale sin devolver error
	if err != nil {
		cr.log.Warn("cache encode failed", "key", key, "error", err)
		return
	}

	// Crea el contexto para redis y difiere el cierre
	rctx, cancel := context.WithTimeout(ctx, redisTimeout)
	defer cancel()

	// Guarda el valor en Redis con dicha clave, si la escritura falla, deja constancia en el log y sigue
	if err := cr.rdb.Set(rctx, key, data, cr.ttl).Err(); err != nil {
		cr.log.Warn("cache write failed", "key", key, "error", err)
	}
}
