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
- Español: comentarios en todos los ficheros (Go, SQL, `.proto`, YAML, Makefile). Breves: una línea por cada decisión no obvia. Nada de explicar lo evidente.

## Repositorio

- `gen/` se genera con `make proto-gen` y nunca se edita a mano.
- Un servicio nunca importa `services/<otro>/internal/`. `pkg/` es solo infraestructura, nunca dominio.
- Versiones siempre fijadas en imágenes Docker y dependencias. Nunca `latest`.
- Si creas, renombras o eliminas un directorio de primer nivel o un fichero de configuración de la raíz, o cambias cómo se arranca el proyecto (targets de `make`, puertos, requisitos), actualiza `README.md` en el mismo cambio. No añadas detalle por debajo del primer nivel.
