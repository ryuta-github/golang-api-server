package domain

import "context"

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	List(ctx context.Context) ([]Task, error)
	FindByID(ctx context.Context, id string) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id string) error
}
