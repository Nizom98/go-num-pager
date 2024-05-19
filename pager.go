package page

import (
	"context"
	"fmt"
)

//go:generate mockgen -source=pager.go -destination mocks_world_test.go -package page

const (
	defaultNextPageStartAt = 0
	defaultTotalCount      = 1
	defaultPageSize        = 100
)

type (
	Loader[T any] interface {
		Load(ctx context.Context, pageStartAt, pageSize int) (page []T, err error)
	}

	LoaderWithNewTotal[T any] interface {
		Load(ctx context.Context, pageStartAt, pageSize int) (page []T, newTotalCount int, err error)
	}

	Pager[T any] struct {
		// elements count per page.
		pageSize int
		// total pages count.
		// If the nextPageLoaderWithNewTotalCount is set,
		// this value will be updated after each call of nextPageLoaderWithNewTotalCount.
		// 1, 2, 3, ...
		totalCount int
		// next page number, that will be loaded in next call of Next.
		// 0, 1, 2, ..., totalCount - 1.
		nextPageStartAt int
		// loader that loads the next page of elements.
		// nextPageLoader or nextPageLoaderWithNewTotalCount must be set.
		nextPageLoader Loader[T]
		// loader that loads the next page of elements and returns the new total count of elements.
		// nextPageLoaderWithNewTotalCount or nextPageLoader must be set.
		nextPageLoaderWithNewTotalCount LoaderWithNewTotal[T]
	}
)

// New creates a new Pager.
// It requires at least one option: WithLoader or WithLoaderWithNewTotal.
// If you do not know the total pages count, use WithLoaderWithNewTotal.
// If you know the total pages count, use WithLoader.
func New[T any](opts ...Option[T]) (*Pager[T], error) {
	pager := &Pager[T]{
		pageSize:        defaultPageSize,
		totalCount:      defaultTotalCount,
		nextPageStartAt: defaultNextPageStartAt,
	}

	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}

	if pager.nextPageLoader == nil && pager.nextPageLoaderWithNewTotalCount == nil {
		return nil, fmt.Errorf("next page loader is required")
	}
	if pager.nextPageStartAt >= pager.totalCount {
		return nil, fmt.Errorf("next page start position must be less than total count")
	}

	return pager, nil
}

// Next returns the next page of elements.
// It uses LoaderWithNewTotal if it is set, otherwise it uses Loader to load data.
func (p *Pager[T]) Next(ctx context.Context) ([]T, error) {
	if p.nextPageLoaderWithNewTotalCount != nil {
		return p.nextWithNewTotal(ctx)
	}

	if p.IsAllLoaded() {
		return nil, nil
	}

	page, err := p.nextPageLoader.Load(ctx, p.nextPageStartAt, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("page start at %d: %w", p.nextPageStartAt, err)
	}
	p.nextPageStartAt += p.pageSize
	return page, nil
}

// nextWithNewTotal returns the next page of elements.
// It uses LoaderWithNewTotal loader to load page data.
func (p *Pager[T]) nextWithNewTotal(ctx context.Context) ([]T, error) {
	if p.IsAllLoaded() {
		return nil, nil
	}

	page, newTotal, err := p.nextPageLoaderWithNewTotalCount.Load(ctx, p.nextPageStartAt, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("page start at %d: %w", p.nextPageStartAt, err)
	}
	p.nextPageStartAt += p.pageSize
	p.totalCount = newTotal
	return page, nil
}

// All returns all elements from all pages.
func (p *Pager[T]) All(ctx context.Context) ([]T, error) {
	allPages := make([]T, 0, p.totalCount*p.pageSize)
	for !p.IsAllLoaded() {
		page, err := p.Next(ctx)
		if err != nil {
			return allPages, err
		}
		allPages = append(allPages, page...)
	}
	return allPages, nil
}

func (p *Pager[T]) IsAllLoaded() bool {
	return p.nextPageStartAt >= p.totalCount
}
