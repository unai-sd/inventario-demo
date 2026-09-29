// Command migrate applies the pending SQL migrations to DATABASE_URL.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"distribuciones-demo.example/inventario/internal/migrate"
	"distribuciones-demo.example/inventario/migrations"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	all, err := migrate.Load(migrations.FS)
	if err != nil {
		log.Fatal(err)
	}
	applied, err := migrate.Up(ctx, conn, all)
	for _, name := range applied {
		fmt.Println("applied", name)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("database up to date (%d migrations, %d applied now)\n", len(all), len(applied))
}
