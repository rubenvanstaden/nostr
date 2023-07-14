package core

import (
	"context"

	"github.com/rubenvanstaden/nostr"
)

type Repository interface {
	Store(ctx context.Context, e *nostr.Event) error
	FindByIdPrefix(ctx context.Context, prefixes []string) ([]nostr.Event, error)
	FindByAuthors(ctx context.Context, authors []string) ([]nostr.Event, error)
	Find(ctx context.Context, id string) (*nostr.Event, error)
}
