package internal

import (
	"context"
	"errors"

	catalogv1 "github.com/cristianrisueo/thebattleroyale1/gen/catalog/v1"
	commonv1 "github.com/cristianrisueo/thebattleroyale1/gen/common/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CatalogHandler actúa como el handler para las peticiones gRPC recibidas
type CatalogHandler struct {
	catalogv1.UnimplementedCatalogServiceServer
	svc Service
}

// Comprueba que CatalogHandler cumple el contrato gRPC
var _ catalogv1.CatalogServiceServer = (*CatalogHandler)(nil)

// NewCatalogHandler crea el handler sobre el service recibido por inyección
func NewCatalogHandler(svc Service) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

//* Implementación del contrato gRPC

// GetStudent devuelve el alumno pedido
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - req: Petición con el id del alumno
// Returns - GetStudentResponse/error: El alumno, o InvalidArgument si el id no es UUID y NotFound si no existe
func (ch *CatalogHandler) GetStudent(ctx context.Context, req *catalogv1.GetStudentRequest) (*catalogv1.GetStudentResponse, error) {
	// Llama al servicio para obtner el alumno por ID
	s, err := ch.svc.GetStudent(ctx, req.GetId())

	// Si el servicio devuelve error traduce el error de nuestro dominio a mensaje protobuf y lo devuelve
	if err != nil {
		return nil, toStatus(err)
	}

	// Si todo va bien devuelve el alumno traducido de nuestro dominio a mensaje protobuf
	return &catalogv1.GetStudentResponse{Student: studentToProto(s)}, nil
}

// ListStudents devuelve todos los alumnos ordenados por nombre
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - _: La petición no tiene campos
// Returns - ListStudentsResponse/error: Los alumnos, o error si falla la consulta
func (ch *CatalogHandler) ListStudents(ctx context.Context, _ *catalogv1.ListStudentsRequest) (*catalogv1.ListStudentsResponse, error) {
	// Llama al servicio para obtener los alumnos ordenados por nombre
	students, err := ch.svc.ListStudents(ctx)

	// Si el servicio devuelve error traduce el error de nuestro dominio a mensaje protobuf y lo devuelve
	if err != nil {
		return nil, toStatus(err)
	}

	// Crea un array del tamaño de los estudiantes recibidos de tipo mensaje students protobuf
	out := make([]*commonv1.Student, 0, len(students))

	// Añade cada estudainte traducido de nuestro dominio a tipo protobuf
	for _, s := range students {
		out = append(out, studentToProto(s))
	}

	// Si todo va bien devuelve el array de estudiantes
	return &catalogv1.ListStudentsResponse{Students: out}, nil
}

// GetWeapon devuelve el arma pedida
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - req: Petición con el id del arma
// Returns - GetWeaponResponse/error: El arma, o InvalidArgument si el id no es UUID y NotFound si no existe
func (ch *CatalogHandler) GetWeapon(ctx context.Context, req *catalogv1.GetWeaponRequest) (*catalogv1.GetWeaponResponse, error) {
	// Llama al servicio para obtener el arma por ID
	w, err := ch.svc.GetWeapon(ctx, req.GetId())

	// Si el servicio devuelve error traduce el error de nuestro dominio a mensaje protobuf y lo devuelve
	if err != nil {
		return nil, toStatus(err)
	}

	// Si todo va bien devuelve el arma traducida de nuestro dominio a mensaje protobuf
	return &catalogv1.GetWeaponResponse{Weapon: weaponToProto(w)}, nil
}

// ListWeapons devuelve todas las armas ordenadas por nombre
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - _: La petición no tiene campos
// Returns - ListWeaponsResponse/error: Las armas, o error si falla la consulta
func (ch *CatalogHandler) ListWeapons(ctx context.Context, _ *catalogv1.ListWeaponsRequest) (*catalogv1.ListWeaponsResponse, error) {
	// Llama al servicio para obtener las armas ordenadas por nombre
	weapons, err := ch.svc.ListWeapons(ctx)

	// Si el servicio devuelve error traduce el error de nuestro dominio a mensaje protobuf y lo devuelve
	if err != nil {
		return nil, toStatus(err)
	}

	// Crea un array del tamaño de las armas recibidas de tipo mensaje weapons protobuf
	out := make([]*commonv1.Weapon, 0, len(weapons))

	// Añade cada arma traducida de nuestro dominio a tipo protobuf
	for _, w := range weapons {
		out = append(out, weaponToProto(w))
	}

	// Si todo va bien devuelve el array de armas
	return &catalogv1.ListWeaponsResponse{Weapons: out}, nil
}

// GetLocation devuelve la localización pedida
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - req: Petición con el id de la localización
// Returns - GetLocationResponse/error: La localización, o InvalidArgument si el id no es UUID y NotFound si no existe
func (ch *CatalogHandler) GetLocation(ctx context.Context, req *catalogv1.GetLocationRequest) (*catalogv1.GetLocationResponse, error) {
	// Llama al servicio para obtener la localización por ID
	l, err := ch.svc.GetLocation(ctx, req.GetId())

	// Si el servicio devuelve error traduce el error de nuestro dominio a mensaje protobuf y lo devuelve
	if err != nil {
		return nil, toStatus(err)
	}

	// Si todo va bien devuelve la localización traducida de nuestro dominio a mensaje protobuf
	return &catalogv1.GetLocationResponse{Location: locationToProto(l)}, nil
}

// ListLocations devuelve todas las localizaciones ordenadas por nombre
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - _: La petición no tiene campos
// Returns - ListLocationsResponse/error: Las localizaciones, o error si falla la consulta
func (ch *CatalogHandler) ListLocations(ctx context.Context, _ *catalogv1.ListLocationsRequest) (*catalogv1.ListLocationsResponse, error) {
	// Llama al servicio para obtener las localizaciones ordenadas por nombre
	locations, err := ch.svc.ListLocations(ctx)

	// Si el servicio devuelve error traduce el error de nuestro dominio a mensaje protobuf y lo devuelve
	if err != nil {
		return nil, toStatus(err)
	}

	// Crea un array del tamaño de las localizaciones recibidas de tipo mensaje locations protobuf
	out := make([]*commonv1.Location, 0, len(locations))

	// Añade cada localización traducida de nuestro dominio a tipo protobuf
	for _, l := range locations {
		out = append(out, locationToProto(l))
	}

	// Si todo va bien devuelve el array de localizaciones
	return &catalogv1.ListLocationsResponse{Locations: out}, nil
}

//* Funciones auxiliares para traducción de errores y conversión a protobuf

// toStatus traduce un error de dominio a un status gRPC
//
// Param - err: Error devuelto por el servicio
// Returns - error: InvalidArgument para ErrInvalidID, NotFound para ErrNotFound e Internal para el resto
func toStatus(err error) error {
	switch {
	// El id no tenía formato de UUID: culpa del cliente
	case errors.Is(err, ErrInvalidID):
		return status.Error(codes.InvalidArgument, "invalid id")

	// No existe ninguna fila con ese id
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, "not found")

	// Usa un mensaje genérico para no filtrar detalles de Postgres al cliente
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

// studentToProto convierte un alumno de dominio a su mensaje protobuf
//
// Param - s: Alumno del dominio
// Returns - commonv1.Student: El mismo alumno como mensaje protobuf
func studentToProto(s Student) *commonv1.Student {
	return &commonv1.Student{
		Id:       s.ID,
		Name:     s.Name,
		Strength: s.Strength,
		Agility:  s.Agility,
		Stamina:  s.Stamina,
	}
}

// weaponToProto convierte un arma del dominio a su mensaje protobuf
//
// Param - w: Arma del dominio
// Returns - commonv1.Weapon: La misma arma como mensaje protobuf
func weaponToProto(w Weapon) *commonv1.Weapon {
	return &commonv1.Weapon{
		Id:       w.ID,
		Name:     w.Name,
		Damage:   w.Damage,
		Accuracy: w.Accuracy,
	}
}

// locationToProto convierte una localización del dominio a su mensaje protobuf
//
// Param - l: Localización del dominio
// Returns - commonv1.Location: El misma localización como mensaje protobuf
func locationToProto(l Location) *commonv1.Location {
	return &commonv1.Location{
		Id:          l.ID,
		Name:        l.Name,
		Description: l.Description,
	}
}
