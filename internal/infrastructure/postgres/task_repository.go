package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ryuta-github/golang-api-server/internal/domain"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

const taskColumns = `id::text, title, description, completed, created_at, updated_at`

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO tasks (title, description)
		VALUES ($1, $2)
		RETURNING `+taskColumns, task.Title, task.Description)
	return scanTask(row, task)
}

func (r *TaskRepository) List(ctx context.Context) ([]domain.Task, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+taskColumns+` FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, translateError(err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt); err != nil {
			return nil, translateError(err)
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, translateError(err)
	}
	return tasks, nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	var task domain.Task
	err := scanTask(r.pool.QueryRow(ctx, `SELECT `+taskColumns+` FROM tasks WHERE id = $1`, id), &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	return scanTask(r.pool.QueryRow(ctx, `
		UPDATE tasks
		SET title = $2, description = $3, completed = $4, updated_at = now()
		WHERE id = $1
		RETURNING `+taskColumns,
		task.ID, task.Title, task.Description, task.Completed), task)
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return translateError(err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner, task *domain.Task) error {
	err := row.Scan(&task.ID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.UpdatedAt)
	return translateError(err)
}

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "22P02": // invalid_text_representation (for example, a malformed UUID)
			return domain.ErrValidation
		case "23505": // unique_violation
			return domain.ErrConflict
		}
	}
	return err
}
