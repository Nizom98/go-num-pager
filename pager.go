package page

import (
	"context"
	"fmt"
)

//go:generate mockgen -source=pager.go -destination mocks_world_test.go -package page

const (
	defaultNextPageNum     = 1
	defaultTotalPagesCount = 1
	defaultPageSize        = 100
)

type (
	Loader[T any] interface {
		Load(ctx context.Context, pageNum, pageSize int) (page []T, err error)
	}

	LoaderWithNewTotal[T any] interface {
		Load(ctx context.Context, pageNum, pageSize int) (page []T, newTotalCount int, err error)
	}

	Pager[T any] struct {
		// elements count per page.
		pageSize int
		// total pages count.
		// If the nextPageLoaderWithNewTotalCount is set,
		// this value will be updated after each call of nextPageLoaderWithNewTotalCount.
		totalPagesCount int
		// next page number, that will be loaded in next call of Next.
		// 1, 2, 3, ..., totalPagesCount.
		nextPageNum int
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
		totalPagesCount: defaultTotalPagesCount,
		nextPageNum:     defaultNextPageNum,
	}

	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}

	if pager.nextPageLoader == nil && pager.nextPageLoaderWithNewTotalCount == nil {
		return nil, fmt.Errorf("next page loader is required")
	}
	if pager.nextPageNum > pager.totalPagesCount {
		return nil, fmt.Errorf("next page number must be less or equal to total pages count")
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

	page, err := p.nextPageLoader.Load(ctx, p.nextPageNum, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("page %d: %w", p.nextPageNum, err)
	}
	p.nextPageNum++
	return page, nil
}

// nextWithNewTotal returns the next page of elements.
// It uses LoaderWithNewTotal loader to load page data.
func (p *Pager[T]) nextWithNewTotal(ctx context.Context) ([]T, error) {
	if p.IsAllLoaded() {
		return nil, nil
	}

	page, newTotal, err := p.nextPageLoaderWithNewTotalCount.Load(ctx, p.nextPageNum, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("page %d: %w", p.nextPageNum, err)
	}
	p.nextPageNum++
	p.totalPagesCount = TotalCountToTotalPagesCount(newTotal, p.pageSize)
	return page, nil
}

// All returns all elements from all pages.
func (p *Pager[T]) All(ctx context.Context) ([]T, error) {
	allPages := make([]T, 0, p.totalPagesCount*p.pageSize)
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
	return p.nextPageNum > p.totalPagesCount
}

// TotalCountToTotalPagesCount calculates the total pages count by total count and page size.
func TotalCountToTotalPagesCount(totalCount, pageSize int) int {
	if pageSize == 0 {
		return 0
	}
	return (totalCount + pageSize - 1) / pageSize
}
