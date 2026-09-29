package migrate

import (
	"testing"
	"testing/fstest"

	"distribuciones-demo.example/inventario/migrations"
)

func TestLoadSortsByVersion(t *testing.T) {
	fsys := fstest.MapFS{
		"0002_second.sql": {Data: []byte("SELECT 2")},
		"0001_first.sql":  {Data: []byte("SELECT 1")},
		"README.md":       {Data: []byte("ignored")},
	}
	got, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "first" || got[1].Version != 2 {
		t.Fatalf("unexpected order: %+v", got)
	}
}

func TestLoadRejectsBadNamesAndDuplicates(t *testing.T) {
	if _, err := Load(fstest.MapFS{"init.sql": {Data: []byte("")}}); err == nil {
		t.Error("expected error for unnumbered file")
	}
	dup := fstest.MapFS{
		"0001_a.sql": {Data: []byte("")},
		"0001_b.sql": {Data: []byte("")},
	}
	if _, err := Load(dup); err == nil {
		t.Error("expected error for duplicate version")
	}
}

func TestPending(t *testing.T) {
	all := []Migration{{Version: 1}, {Version: 2}, {Version: 3}}
	got := Pending(all, map[int]bool{1: true, 3: true})
	if len(got) != 1 || got[0].Version != 2 {
		t.Fatalf("unexpected pending: %+v", got)
	}
}

func TestEmbeddedMigrationsAreValid(t *testing.T) {
	got, err := Load(migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	for i, m := range got {
		if m.Version != i+1 {
			t.Fatalf("migration versions must be consecutive from 0001, got %04d at position %d", m.Version, i)
		}
	}
}
