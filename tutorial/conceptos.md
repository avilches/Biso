# Conceptos

`biso` es la herramienta con la que un agente automático, y la persona que trabaja con él, llevan
las tareas de un proyecto. Antes de ver un solo comando conviene tener claros cinco conceptos, porque
el resto del tutorial da por hecho que ya los conoces.

## El tablero

Un tablero es el conjunto de tareas de un proyecto junto con su configuración: qué estados existen,
qué tipos de tarea hay, qué prioridades se pueden usar y quién es, si alguien lo es, la identidad de
quien trabaja en él. Esa configuración no es un detalle aparte: es lo que hace que un valor tenga
sentido o no. Un tablero no acepta cualquier texto en cualquier campo, sino solo los valores de su
propio vocabulario.

De ahí sale la regla que gobierna todo lo demás: un valor que el tablero no reconoce es siempre un
error, tanto si se está escribiendo como si se está preguntando por él, y con la misma exigencia en
los dos casos. Preguntar por un estado que ese tablero no tiene no devuelve una lista vacía, devuelve
un error. Como consecuencia, cuando una pregunta sí devuelve una lista vacía es porque de verdad no
hay nada que cumpla lo pedido, y eso es información, no un fallo silencioso.

## Los estados

Cada tarea tiene un estado, y el estado es uno de los que ese tablero tiene configurados: no hay una
lista fija de estados válida para todos los proyectos, cada tablero declara los suyos. Lo único que no
se puede configurar libremente es que, de entre esos estados, tres papeles queden siempre cubiertos,
cada uno por un estado distinto:

- El estado **inicial**, donde nace una tarea nueva.
- El estado **activo**, el que se escribe cuando alguien coge una tarea para trabajar en ella.
- El estado **terminal**, el que marca que una tarea está acabada.

Un papel es la función que cumple un estado, no su nombre. Dos tableros pueden llamar de forma
distinta al estado inicial y aun así los dos tienen uno, porque el papel es obligatorio aunque la
palabra con la que se llama no lo sea.

## Criterios de aceptación y definición de hecho

Una tarea puede llevar dos listas de comprobación independientes, con la misma forma: cada elemento
tiene un texto y puede estar marcado o no, y cada elemento nace con una clave numérica propia que no
cambia aunque se quiten otros elementos de la lista. Son dos listas, no una, y se llevan la cuenta por
separado.

La primera son los **criterios de aceptación**: cómo se sabe que el trabajo de esa tarea en concreto
funciona. La segunda es la **definición de hecho**: lo que tiene que ser cierto antes de dar la tarea
por cerrada, más allá de si el resultado funciona, como que otra persona la haya revisado. Ninguna de
las dos es obligatoria, y una tarea puede llegar a su estado terminal con elementos sin marcar en
cualquiera de las dos, porque `biso` avisa de eso pero no lo impide por defecto.

## La urgencia

La urgencia es un número que no se guarda en ningún sitio: se calcula cada vez que se lee la tarea, a
partir de datos que sí están guardados, como la prioridad, si la tarea está activa, si bloquea o la
bloquea otra tarea, si tiene fecha límite cercana, si tiene criterios de aceptación y cuánto tiempo
lleva abierta. Por eso se dice que la urgencia es **derivada**: nadie la escribe directamente, y su
valor de ahora mismo puede no ser el de dentro de un minuto, aunque nadie haya tocado la tarea,
sencillamente porque el tiempo pasó o porque cambió alguna otra tarea de la que depende. Una tarea en
su estado terminal tiene urgencia cero siempre, sin más cálculo.

## La identidad declarada

Un tablero puede tener configurada una identidad, el texto que dice quién eres tú cuando llamas a
`biso`. No es obligatoria: un tablero funciona perfectamente sin que nadie la declare. Lo que cambia
es que algunas operaciones necesitan saber quién eres para tener sentido, como pedir solo las tareas
que son tuyas, quedarte automáticamente asignado una tarea al empezarla, firmar un comentario con tu
nombre por defecto, o dejar una pregunta abierta a la espera de que alguien responda. Si intentas
cualquiera de esas cosas sin identidad declarada, `biso` no adivina ni asume nada: lo dice como un
error explícito. El resto del tablero, lo que no necesita saber quién eres, sigue funcionando igual
con identidad declarada o sin ella.
