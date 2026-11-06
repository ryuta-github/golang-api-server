//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ryuta-github/golang-api-server/internal/domain"
)

func TestTaskRepositoryCRUD(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, "TRUNCATE tasks"); err != nil {
		t.Fatal(err)
	}

	repository := NewTaskRepository(pool)
	if _, err := repository.FindByID(ctx, "not-a-uuid"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("FindByID() malformed ID error = %v, want ErrValidation", err)
	}

	task, _ := domain.NewTask("integration task", "details")
	if err := repository.Create(ctx, task); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if task.ID == "" || task.CreatedAt.IsZero() {
		t.Fatalf("Create() task = %+v", task)
	}

	tasks, err := repository.List(ctx)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("List() tasks = %+v, error = %v", tasks, err)
	}

	task.Completed = true
	if err := repository.Update(ctx, task); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	found, err := repository.FindByID(ctx, task.ID)
	if err != nil || !found.Completed {
		t.Fatalf("FindByID() task = %+v, error = %v", found, err)
	}

	if err := repository.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := repository.FindByID(ctx, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("FindByID() error = %v, want ErrNotFound", err)
	}
}
