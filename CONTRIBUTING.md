# Cómo contribuir

## Flujo de ramas: tronco + ramas de mantenimiento

| Rama | Para qué sirve |
| --- | --- |
| `main` | Tronco. Todo el trabajo nuevo entra aquí mediante pull request. Es la rama por defecto. |
| `release/1.x` | Mantenimiento de la serie 1.x. De aquí salen las etiquetas `v1.Y.Z`. Solo recibe correcciones traídas desde `main` y la sincronización de cada versión menor. |

Reglas:

1. Crea tu rama desde `main` con un nombre corto: `feat/<slug>`, `fix/<slug>`, `docs/<slug>`
   o `chore/<slug>`.
2. Abre la PR contra `main`. Se integra con **squash merge**: el título de la PR será el
   mensaje del commit, así que usa [Conventional Commits](https://www.conventionalcommits.org/es/v1.0.0/)
   en español (`feat: ...`, `fix: ...`).
3. Nunca se abre una PR de funcionalidad contra `release/1.x`.

### Backports (correcciones a una versión publicada)

Las correcciones **siempre llegan primero a `main`**. Si también deben llegar a la serie 1.x:

```bash
git switch -c backport/<numero-pr>-1.x origin/release/1.x
git cherry-pick -x <sha-del-squash-en-main>
git push -u origin HEAD
```

Abre la PR contra `release/1.x` con la etiqueta **`backport`** y el título del commit
original seguido de `(backport #<numero-pr>)`. Se integra con squash. El trabajo
`backport-label` de la integración continua falla si la PR no lleva la etiqueta.

### Publicar una versión

- **Parche (`v1.Y.Z`)**: tras integrar los backports en `release/1.x`, se etiqueta el último
  commit de `release/1.x` con una etiqueta anotada.
- **Versión menor (`v1.Y.0`)**: se abre una PR de `main` a `release/1.x` con la etiqueta
  **`release-sync`**, se integra con **merge commit** (no squash, para conservar los commits
  de `main`) y se etiqueta el merge.
- Al empujar la etiqueta se ejecuta el flujo `release`, que compila y simula la publicación.
- Las etiquetas solo se crean sobre commits de `release/*`; el flujo lo comprueba.

## Antes de abrir la PR

```bash
make lint
make test
make db-up && make test-integration
```

La integración continua ejecuta `lint`, `test` e `integration` (con un PostgreSQL 16 de
servicio) en cada PR contra `main` o `release/*`.

## Migraciones

Un cambio de esquema es un fichero nuevo en `migrations/` con el siguiente número. No se
editan migraciones ya integradas en `main`.
