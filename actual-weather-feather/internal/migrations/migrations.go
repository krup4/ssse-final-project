package migrations

import (
    "context"
    "embed"
    "fmt"
    "sort"
    "strings"

    "github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var migrationFiles embed.FS

func Apply(ctx context.Context, db *pgxpool.Pool) error {
    entries, err := migrationFiles.ReadDir(".")
    if err != nil {
        return fmt.Errorf("read migrations: %w", err)
    }

    var names []string
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        if strings.HasSuffix(e.Name(), ".sql") {
            names = append(names, e.Name())
        }
    }
    sort.Strings(names)

    for _, name := range names {
        data, err := migrationFiles.ReadFile(name)
        if err != nil {
            return fmt.Errorf("read migration %s: %w", name, err)
        }
        if _, err := db.Exec(ctx, string(data)); err != nil {
            return fmt.Errorf("apply migration %s: %w", name, err)
        }
    }
    return nil
}
