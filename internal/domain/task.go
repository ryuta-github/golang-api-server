package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
	ErrValidation = errors.New("validation failed")
)

type Task struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTask(title, description string) (*Task, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > 200 {
		return nil, ErrValidation
	}
	if len(description) > 2000 {
		return nil, ErrValidation
	}

	return &Task{Title: title, Description: description}, nil
}

func (t *Task) Update(title, description *string, completed *bool) error {
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" || len(trimmed) > 200 {
			return ErrValidation
		}
		t.Title = trimmed
	}
	if description != nil {
		if len(*description) > 2000 {
			return ErrValidation
		}
		t.Description = *description
	}
	if completed != nil {
		t.Completed = *completed
	}
	return nil
}
