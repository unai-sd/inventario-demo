# inventario-demo

API de inventario de almacenes de la distribuidora ficticia **Distribuciones Demo**:
productos, almacenes, movimientos de stock y alertas de stock bajo.

## Requisitos

- Go 1.27 (declarado en `go.mod` con la directiva `toolchain` y en `.tool-versions` para asdf/mise).
- Docker con `docker compose` para levantar PostgreSQL 16.
- `make`.

## Puesta en marcha

```bash
make deps        # descarga los módulos de go.sum
make db-up       # PostgreSQL 16 en localhost:5433 (docker compose)
make migrate     # aplica migrations/*.sql
make seed        # (opcional) almacenes, productos y stock de demostración
make run         # API en http://localhost:8080
```

`DATABASE_URL` y `PORT` tienen valores por defecto en el `Makefile`; ver `.env.example`.
`make db-down` para el contenedor y borra los datos.

## Comandos habituales

| Tarea | Comando |
| --- | --- |
| Lint (`go vet` + `gofmt`) | `make lint` |
| Formatear | `make fmt` |
| Pruebas unitarias (sin base de datos) | `make test` |
| Pruebas de integración (necesitan `make db-up`) | `make test-integration` |
| Compilar binarios en `bin/` | `make build` |

Las pruebas de integración llevan la etiqueta de compilación `integration`, leen
`DATABASE_URL`, aplican las migraciones y vacían las tablas: no las lances contra una base
de datos con datos que quieras conservar.

## Endpoints

| Método | Ruta | Descripción |
| --- | --- | --- |
| GET | `/healthz` | Estado y versión |
| GET, POST | `/products` | Listar y dar de alta productos |
| GET | `/products/{id}` | Detalle de un producto |
| GET, POST | `/warehouses` | Listar y dar de alta almacenes |
| GET, POST | `/movements` | Libro de movimientos (filtro `product_id`) y registrar entradas (`in`), salidas (`out`) o ajustes de inventario (`adjust`) |
| GET | `/stock` | Stock por producto y almacén (filtro `warehouse_id`) |
| GET | `/alerts/low-stock` | Productos con stock total igual o inferior a su mínimo |
| POST | `/transfers` | Traspaso de stock entre dos almacenes |

Ejemplo:

```bash
curl -X POST localhost:8080/movements \
  -d '{"product_id":1,"warehouse_id":1,"kind":"in","quantity":24,"reason":"albarán 1001"}'
```

## Documentación

- [Arquitectura y modelo de datos](docs/arquitectura.md)
- [Cómo contribuir y flujo de ramas](CONTRIBUTING.md)
