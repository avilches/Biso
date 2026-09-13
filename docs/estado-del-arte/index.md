# Estado del arte: gestores de tareas para agentes, y por qué fallan

Investigado el 2026-09-07 para tomar la decisión de persistencia de `biso`
([diseño](../superpowers/specs/2026-09-07-persistencia-design.md)).

La primera mitad es un inventario de las herramientas del espacio, y se cierra con una comparación
aparte: cómo se le dice a cada una que trabaje con un almacén distinto del que encuentra sola, donde
entran también herramientas de fuera del espacio. La segunda es un catálogo de
problemas: cada uno con quién lo sufre, la evidencia con su enlace, y la respuesta de `biso`. La
segunda mitad es la que sirve para explicar la herramienta a alguien que viene de otra.

## Aviso de método

Lo que sigue se apoya en repositorios de GitHub (código, issues, documentación), en Hacker News, y en
documentación oficial. **No hay ninguna cita de Reddit**: el dominio rechaza las descargas
automatizadas y también fallaron las vías alternativas, así que las opiniones de r/ClaudeAI y
r/ChatGPTCoding no están representadas.

Cinco cosas concretas quedaron sin verificar del todo, y se marcan también donde aparecen:

1. **La cita de Steve Yegge** sobre que los conflictos del JSONL de Beads los resuelve la propia IA al
   fusionar. El artículo de Medium devolvió un error 403 y nunca se leyó directamente.
2. **Las versiones exactas del fallo de WAL** de SQLite entre procesos concurrentes (3.7.0 a 3.51.2,
   corregido en 3.51.3), que salen de un resumen de la página oficial y no de su registro de cambios.
3. **El número de estrellas de Spec Kit**, leído de la página y no de la interfaz de programación.
4. **Las tasas de conflicto del 19,8% y del 41,7%** entre pull requests de agentes, que vienen de un
   único artículo y no se pudieron cruzar con una segunda fuente.
5. **La corrupción de SQLite específicamente sobre iCloud Drive o Dropbox por su nombre.** La
   documentación oficial habla de sistemas de ficheros en red en general; aplicarlo a esas carpetas es
   una extrapolación razonable, no un hecho citado.

## Las páginas

- [Parte 1. Las herramientas](herramientas.md)
- [Parte 2. El catálogo de problemas](catalogo-de-problemas.md)
- [Parte 3. Lo que la gente quiere conservar, y lo que rechaza](lo-que-se-conserva.md)
