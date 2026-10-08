## Qué cambia

<!-- Una o dos frases. El título de la PR será el mensaje del commit (squash). -->

## Tipo

- [ ] Funcionalidad o corrección contra `main`
- [ ] Backport a `release/1.x` (etiqueta `backport`; enlaza la PR original: #)
- [ ] Sincronización de versión menor `main` → `release/1.x` (etiqueta `release-sync`)

## Comprobaciones

- [ ] `make lint` y `make test` en verde
- [ ] `make test-integration` en verde si toca `internal/store` o `migrations/`
- [ ] Migración nueva (no editada) si cambia el esquema
- [ ] ¿Debe llevarse a `release/1.x`? Indícalo para preparar el backport
