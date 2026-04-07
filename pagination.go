package signdocsbrasil

import (
	"context"
)

// PageIterator provides a generic way to iterate through paginated results.
// It uses the nextToken pattern used by the SignDocs API.
//
// Usage:
//
//	iter := client.Transactions.ListAutoPaginate(&signdocsbrasil.TransactionListParams{
//	    Status: signdocsbrasil.TransactionStatusCompleted,
//	})
//	for iter.Next(ctx) {
//	    tx := iter.Value()
//	    // process tx
//	}
//	if err := iter.Err(); err != nil {
//	    // handle error
//	}
type PageIterator[T any] struct {
	fetch   func(ctx context.Context, nextToken string) ([]T, string, error)
	items   []T
	pos     int
	token   string
	started bool
	done    bool
	err     error
}

// newPageIterator creates a new PageIterator. The fetch function should return
// the items for the current page, the next page token (empty string if no more pages),
// and any error.
func newPageIterator[T any](fetch func(ctx context.Context, nextToken string) ([]T, string, error)) *PageIterator[T] {
	return &PageIterator[T]{
		fetch: fetch,
	}
}

// Next advances the iterator to the next item. It returns false when iteration
// is complete or an error has occurred. Call Value() to get the current item
// and Err() to check for errors after iteration.
func (p *PageIterator[T]) Next(ctx context.Context) bool {
	if p.done || p.err != nil {
		return false
	}

	// If we have items remaining in the current page, advance
	if p.started && p.pos < len(p.items)-1 {
		p.pos++
		return true
	}

	// If we have exhausted the current page and there is no next token, we are done
	if p.started && p.token == "" {
		p.done = true
		return false
	}

	// Fetch the next page
	items, nextToken, err := p.fetch(ctx, p.token)
	if err != nil {
		p.err = err
		return false
	}

	p.started = true
	p.items = items
	p.token = nextToken
	p.pos = 0

	if len(items) == 0 {
		p.done = true
		return false
	}

	return true
}

// Value returns the current item. It is only valid after Next() returns true.
func (p *PageIterator[T]) Value() T {
	return p.items[p.pos]
}

// Err returns the error that caused iteration to stop, if any.
func (p *PageIterator[T]) Err() error {
	return p.err
}
