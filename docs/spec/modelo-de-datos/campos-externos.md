# Los campos externos

`ext` es un `map<string,string>` de clave a texto para guardar cualquier valor que el usuario quiera
asociar a la tarea, bajo una clave elegida por él. Guardar la identidad de la tarea en otro sistema,
como una tarjeta de Trello, es un ejemplo de uso, no la definición del campo. La regla es la
siguiente:

- El tablero **declara** en su configuración qué claves admite, en la lista `extensions`.
- Escribir una clave declarada funciona: `biso set MYP-1 --ext trello.card=5f2a8c1e3b9d4a7f`.
- Escribir una clave no declarada es error 3:
  ```
  error: unknown extension key: "jira.key"
         declared keys on this board: trello.card, github.issue
  ```
- Una tarea que ya guarda una clave que la configuración no declara **no se lee en silencio ni se
  reescribe perdiéndola**: se aplica la regla de ["Qué pasa con un dato que no se puede interpretar"](../garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar), y `biso doctor` la reporta.
