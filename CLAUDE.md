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

Se escriben para quien lea el código dentro de seis meses sin recordar por qué se hizo así. Explican el porqué y el contrato, nunca lo que la línea ya dice.

**En paquetes, tipos y funciones exportadas** pueden ocupar varias líneas y deben cubrir, cuando aplique:

- Qué problema resuelve la pieza y por qué existe aquí y no en otro sitio.
- El contrato con quien la usa: si bloquea, qué significa cada valor devuelto, en qué orden hay que llamar a las cosas.
- Las trampas que no se ven leyendo el código: lo que provoca un pánico, lo que puede colgarse, lo que se queda abierto si falta una llamada.

**Dentro de las funciones**, una única línea física por bloque de lógica, encima del bloque. Comenta cada bloque con lo que aporta o con la consecuencia de esa línea, no con su traducción al castellano. En un `select` o un `switch`, el comentario va en la rama que lo necesita, no encima del bloque entero.

**Nunca**: comentarios que repiten la línea (`// incrementa i`), bloques tipo Parámetros/Retorno/Ejemplo, ni comentarios de relleno en líneas triviales.

## Repositorio

- `gen/` se genera con `make proto-gen` y nunca se edita a mano.
- Un servicio nunca importa `services/<otro>/internal/`. `pkg/` es solo infraestructura, nunca dominio.
- Versiones siempre fijadas en imágenes Docker y dependencias. Nunca `latest`.
- Si creas, renombras o eliminas un directorio de primer nivel o un fichero de configuración de la raíz, o cambias cómo se arranca el proyecto (targets de `make`, puertos, requisitos), actualiza `README.md` en el mismo cambio. No añadas detalle por debajo del primer nivel.
