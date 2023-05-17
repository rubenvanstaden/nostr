package core

import "context"

type Repository interface {
	Store(ctx context.Context, e *Event) error
	Find(ctx context.Context, id EventId) (*Event, error)
}
