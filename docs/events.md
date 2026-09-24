# Catálogo de eventos

Contrato de los eventos que circulan por Kafka: qué significa cada uno, qué lleva, dónde se publica y qué garantías da. Los `.proto` implementan lo que dice este documento; si no coinciden, el error está en el `.proto`.

## Convenciones

- **Los eventos son hechos consumados.** Se nombran en pasado (`BattleStarted`) y nunca piden nada a nadie. Las peticiones entre servicios son comandos y van en otros topics.
- **Protobuf, un tipo de mensaje por topic.** Cada topic transporta un sobre con un `oneof` que contiene el evento concreto. `buf breaking` vigila su evolución.
- **Cada evento se entiende solo.** Un consumidor puede empezar a leer a mitad de una batalla, tras un reinicio o un rebalanceo, sin haber visto los eventos anteriores. Por eso los eventos llevan los nombres además de los ids.
- **Los contratos de eventos no importan contratos de otros servicios.** Los datos del catálogo se copian en mensajes propios de `arena.v1`. Si `catalog.v1` cambia, los eventos no cambian.
- **Solo cambios compatibles.** Se pueden añadir campos y añadir variantes al `oneof`. Nunca se reutiliza ni se cambia el número o el tipo de un campo.
- **Los consumidores ignoran lo que no conocen:** campos nuevos y tipos de evento nuevos. Así un productor puede publicar un evento nuevo sin romper a quien no lo espera.
- **Entrega al menos una vez.** Un consumidor puede recibir el mismo evento más de una vez y debe deduplicar por `event_id`.
- **La traza viaja en las cabeceras de Kafka** (`traceparent`, formato W3C), nunca dentro del mensaje.

## Topic `arena.battles`

| Propiedad    | Valor                                   |
| ------------ | --------------------------------------- |
| Productor    | `arena`                                 |
| Consumidores | `narrator` (fase 3), `betting` (fase 6) |
| Mensaje      | `arena.v1.BattleEvent`                  |
| Clave        | `battle_id`                             |
| Particiones  | 3                                       |
| Limpieza     | Por tiempo (`delete`), sin compactar    |

- **Un único topic para todos los eventos de batalla.** El orden solo se garantiza dentro de una partición. Con un topic por tipo, `BattleFinished` podría consumirse antes que el último ataque. El precio es que `betting` recibe también los ataques y los descarta.
- **3 particiones.** Con varias, las batallas se reparten y se ve que cada una se queda entera en la suya. Además, dos instancias de `narrator` tienen particiones que repartirse en el rebalanceo de la fase 3.
- **Sin compactar.** La compactación guarda solo el último mensaje por clave, es decir, el `BattleFinished` de cada batalla. Reproducir una batalla entera exige conservar todos sus eventos.

## Orden

Dentro de una batalla, el orden es siempre este:

`BattleScheduled` → `BattleStarted` → (`AttackResolved`, `StudentEliminated`)\* → `BattleFinished`

- Un `StudentEliminated` va siempre justo detrás del `AttackResolved` que causó la eliminación.
- `sequence` empieza en 1 en `BattleScheduled` y aumenta de uno en uno, sin huecos. Un hueco o una repetición en el consumidor delatan un problema.
- Entre batallas distintas no hay orden garantizado.

## Sobre: `BattleEvent`

| Campo         | Tipo                        | Contenido                                                                    |
| ------------- | --------------------------- | ---------------------------------------------------------------------------- |
| `event_id`    | `string`                    | UUID único del evento. Clave de deduplicación                                |
| `battle_id`   | `string`                    | UUID de la batalla. Coincide con la clave del mensaje                        |
| `sequence`    | `int64`                     | Posición del evento dentro de su batalla, desde 1                            |
| `occurred_at` | `google.protobuf.Timestamp` | Hora real en que ocurrió, en UTC                                             |
| `battle_time` | `google.protobuf.Duration`  | Tiempo simulado desde el inicio de la batalla. Cero antes de `BattleStarted` |
| `payload`     | `oneof`                     | Uno de los eventos de abajo                                                  |

### Mensajes auxiliares

| Mensaje      | Campos                                         | Uso                                                                            |
| ------------ | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `Fighter`    | `student` (`Student`), `weapon` (`Weapon`)     | Participante completo, solo en `BattleScheduled`                               |
| `Student`    | `id`, `name`, `strength`, `agility`, `stamina` | Copia del alumno tal y como estaba al programar la batalla                     |
| `Weapon`     | `id`, `name`, `damage`, `accuracy`             | Copia del arma asignada                                                        |
| `Location`   | `id`, `name`, `description`                    | Copia de la localización                                                       |
| `FighterRef` | `id`, `name`                                   | Referencia a un participante en los eventos posteriores. `id` es el del alumno |

## Eventos

### `BattleScheduled`

`arena` ha fijado la batalla, sus participantes y la localización. Los datos quedan congelados: la batalla se simula con ellos aunque `catalog` cambie después.

| Campo       | Tipo                        | Contenido                                                                          |
| ----------- | --------------------------- | ---------------------------------------------------------------------------------- |
| `location`  | `Location`                  | Dónde se libra la batalla                                                          |
| `fighters`  | `repeated Fighter`          | Participantes con su arma asignada. Al menos 2                                     |
| `starts_at` | `google.protobuf.Timestamp` | Hora real prevista de inicio. Orientativa: el inicio real lo marca `BattleStarted` |

- `betting` abre el mercado de la batalla.
- `narrator` presenta a los participantes.

### `BattleStarted`

La batalla ha empezado. A partir de aquí no se acepta ninguna apuesta. No lleva campos: el sobre ya dice qué batalla y cuándo.

- `betting` cierra el mercado.
- `narrator` anuncia el inicio.

### `AttackResolved`

Un participante ha atacado a otro, con acierto o sin él.

| Campo           | Tipo         | Contenido                               |
| --------------- | ------------ | --------------------------------------- |
| `attacker`      | `FighterRef` | Quién ataca                             |
| `target`        | `FighterRef` | A quién ataca                           |
| `weapon_name`   | `string`     | Arma usada                              |
| `hit`           | `bool`       | Si el ataque acertó                     |
| `damage`        | `int32`      | Vida restada. 0 si falló                |
| `target_health` | `int32`      | Vida del objetivo tras el ataque, 0-100 |

- `narrator` lo narra.

### `StudentEliminated`

Un participante ha quedado fuera. Su vida ha llegado a 0.

| Campo           | Tipo         | Contenido                       |
| --------------- | ------------ | ------------------------------- |
| `student`       | `FighterRef` | Quién queda eliminado           |
| `eliminated_by` | `FighterRef` | Quién lo elimina                |
| `remaining`     | `int32`      | Participantes que siguen en pie |

- `narrator` lo narra.
- En el futuro, las apuestas en vivo.

### `BattleFinished`

La batalla ha terminado con un único superviviente. Siempre hay ganador: los ataques se resuelven de uno en uno, así que no puede haber eliminaciones simultáneas.

| Campo    | Tipo         | Contenido                  |
| -------- | ------------ | -------------------------- |
| `winner` | `FighterRef` | Último participante en pie |

- `betting` liquida las apuestas.
- `narrator` cierra la crónica.

## Fuera por ahora

- `BattleCancelled`: llega en la fase 8, como variante nueva del `oneof`. Los consumidores que no lo conozcan lo ignorarán.
- Apuestas en vivo: no necesitan eventos nuevos, porque `StudentEliminated` ya les da lo que necesitan.
