package usecase

import (
	"context"

	"github.com/ryuta-github/golang-api-server/internal/domain"
)

type Task struct {
	repository domain.TaskRepository
}

func NewTask(repository domain.TaskRepository) *Task {
	return &Task{repository: repository}
}

func (u *Task) Create(ctx context.Context, title, description string) (*domain.Task, error) {
	task, err := domain.NewTask(title, description)
	if err != nil {
		return nil, err
	}
	if err := u.repository.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (u *Task) List(ctx context.Context) ([]domain.Task, error) {
	return u.repository.List(ctx)
}

func (u *Task) Get(ctx context.Context, id string) (*domain.Task, error) {
	if id == "" {
		return nil, domain.ErrValidation
	}
	return u.repository.FindByID(ctx, id)
}

func (u *Task) Update(ctx context.Context, id string, title, description *string, completed *bool) (*domain.Task, error) {
	if id == "" || (title == nil && description == nil && completed == nil) {
		return nil, domain.ErrValidation
	}

	task, err := u.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := task.Update(title, description, completed); err != nil {
		return nil, err
	}
	if err := u.repository.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (u *Task) Delete(ctx context.Context, id string) error {
	if id == "" {
		return domain.ErrValidation
	}
	return u.repository.Delete(ctx, id)
}
