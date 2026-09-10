# `biso`: especificación del CLI de gestión de tareas

Este documento define `biso`, una herramienta de línea de comandos para llevar las tareas de un
proyecto. Está escrito para que alguien implemente el programa entero a partir de él sin preguntar
nada: cada comando trae su firma, su tabla de parámetros, su comportamiento en los casos límite, la
salida literal que imprime, su esquema JSON, sus códigos de salida y el texto exacto de su ayuda.

El destinatario principal de `biso` es un agente automático que trabaja dentro del proyecto. La
salida es predecible, los errores son distinguibles por su código sin leer el mensaje, y ningún
comportamiento depende de dónde se ejecute el programa.

**Qué define este documento y qué no.** Define la interfaz del programa y el modelo de datos lógico de
una tarea, no el porqué de cómo se guardan los datos: esa decisión, con su razonamiento y su evidencia,
vive en `docs/DECISIONES.md` (sección 12), y este documento no la repite. Aquí aparece el mecanismo
solo donde afecta al comportamiento observable, como el fichero puntero `.biso.json` (sección 3.2) o la
base de datos SQLite que `biso doctor` comprueba (sección 10.11); donde no lo afecta, este documento
enuncia el requisito (por ejemplo, que dos procesos simultáneos no puedan asignar el mismo
identificador) y deja el resto del mecanismo fuera de aquí.

El modelo de tarea es compatible con el de Backlog.md, de modo que se puede importar y exportar entre
las dos herramientas sin perder campos. **La compatibilidad es de modelo de datos y no de formato de
fichero**: los campos se corresponden uno a uno, pero `biso` no lee ni escribe los ficheros Markdown de
esa herramienta, y no hay ninguna intención de que lo haga (sección 14).

**Convención de idioma.** La prosa de este documento va en español. Todo lo que es interfaz del
programa (nombres de comando, banderas, textos de ayuda, mensajes de error, claves JSON y claves de
configuración) va en inglés, porque es lo que la persona o el agente que usa el programa lee y
escribe.


Estas reglas valen para todos los comandos y no se repiten en cada uno. Un comando solo las menciona
cuando se aparta de ellas, y ninguno lo hace salvo donde se diga.

