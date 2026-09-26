---
name: tareas-autocontenidas
description: Cada tarea del tablero se sostiene sola y su objetivo es la herramienta, no migrar tareas de Backlog.md
metadata:
  node_type: memory
  type: feedback
---

Cada tarea del tablero de este proyecto es única e individual: ni su descripción ni sus criterios de
aceptación pueden apoyarse en datos que existan fuera de ella. No citan otras tareas (ni para
depender de ellas ni para repartirse el trabajo), ni mediciones sobre tableros concretos de esta
máquina, ni rutas de la máquina, ni el importador o el exportador de Backlog.md. Quien la retome
tiene que poder entenderla y cerrarla sin haber visto ninguna conversación ni ninguna otra tarea.

El objetivo del proyecto es una herramienta buena, con su especificación y su código. No es migrar
las tareas de Backlog.md a este repositorio. Los tableros de este repositorio y los de otros
proyectos de `~/Hub` sirven para analizar cómo se usa Backlog.md, y esa evidencia va en las páginas
de `docs/decisiones/`, donde se cita como análisis, nunca como premisa en el texto de una tarea.

Se aprendió el 2026-09-21, al preparar las tareas de etiquetas con ámbito, de orden manual y de la
retirada de `ext`: sus descripciones traían mediciones sobre 535 tareas de seis tableros, remitían al
importador y se apoyaban unas en otras, y hubo que reescribirlas.

Cómo se aplica al crear o editar una tarea:

- La descripción dice por qué existe: el problema y la decisión, con un enlace a la entrada de
  `docs/decisiones/` que la recoge cuando la hay.
- Una tarea que nace de una decisión de diseño lleva en sus criterios todo el recorrido: la decisión
  escrita, la especificación, el código con sus pruebas, `estado-de-implementacion.md` con
  `make check` en verde, y una revisión adversarial independiente cuyos hallazgos se resuelven o se
  descartan por escrito con su razón.
- Las notas de la tarea recogen el estado real al cerrar la sesión: qué está hecho, qué falta y qué
  detalles ya decididos no hay que volver a discutir.
- No se usan las dependencias del tablero para acoplar una tarea a otra por comodidad.
