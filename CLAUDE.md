# CLAUDE.md

Monorepo de aprendizaje de microservicios en Go. El diseño se decide fuera de aquí: implementa solo lo que pide cada prompt.

## Forma de trabajo

- Si algo no está especificado o ves un problema en lo que se pide, para y pregunta.
- Si un comando falla, para y muestra el error. No cambies otros ficheros para arreglarlo.
- No añadas dependencias que el prompt no nombre.
- No hagas commits.
- Al terminar, lista los ficheros tocados y la salida de cada comando ejecutado.

## Idioma

- Inglés: código, identificadores, datos, logs y commits.
- Español: comentarios en todos los ficheros (Go, SQL, `.proto`, YAML, Makefile).

## Comentarios

Los ficheros de `services/catalog/` son la referencia: cualquier código nuevo se comenta como ellos.

**Funciones y métodos exportados:**

```go
// GetStudent devuelve el alumno con ese id
//
// Param - ctx: Contexto de la llamada gRPC, cancelado si el cliente se va o vence su deadline
// Param - id: UUID ya validado por el servicio
// Returns - Student/error: El alumno, o ErrNotFound si no existe ninguno con ese id
```

- Primera línea: qué hace, en tercera persona y sin punto final.
- Línea de comentario vacía.
- Una línea `Param - nombre: Descripción` por parámetro, en el orden de la firma, empezando en mayúscula. Los parámetros ignorados se documentan igual (`Param - _: La petición no tiene campos`).
- Una línea `Returns - Tipo/error: Descripción` con qué se devuelve cuando va bien y qué errores concretos salen cuando no.
- Cada línea explica el contrato (qué formato se espera, si puede llegar vacío, qué errores salen y cuándo), nunca el tipo que ya está en la firma.
- Separación con espacios, nunca tabuladores.

**Una sola línea**, misma forma, para todo lo demás: tipos, interfaces, structs, constructores, constantes, bloques de errores, métodos declarados dentro de una interfaz, aserciones de tipo (`var _ Service = ...`) y funciones auxiliares privadas.

**Dentro de las funciones:** una línea encima de cada bloque de lógica, separando los bloques con línea en blanco. La llamada y su `if err != nil` son dos bloques distintos, cada uno con el suyo. En un `switch` o un `select`, el comentario va en cada rama, no encima del bloque.

**Secciones:** dentro de un fichero largo, `//* Nombre de la sección` separa grupos de funciones (por ejemplo, `//* Implementación de la interfaz`).

**Sin comentario de paquete.** Si alguna vez se pone, va pegado a la línea `package`, sin línea en blanco entre medias, o Go no lo reconoce como documentación.

**Nunca:** comentarios que repiten la línea (`// incrementa i`), relleno en líneas triviales, ni párrafos de varias líneas explicando diseño. El porqué de una decisión cabe en la línea del bloque donde se aplica.

## Repositorio

- `gen/` se genera con `make proto-gen` y nunca se edita a mano.
- Un servicio nunca importa `services/<otro>/internal/`. `pkg/` es solo infraestructura, nunca dominio.
- Versiones siempre fijadas en imágenes Docker y dependencias. Nunca `latest`.
- Si creas, renombras o eliminas un directorio de primer nivel o un fichero de configuración de la raíz, o cambias cómo se arranca el proyecto (targets de `make`, puertos, requisitos), actualiza `README.md` en el mismo cambio. No añadas detalle por debajo del primer nivel.
