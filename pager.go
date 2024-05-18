package page

import (
	"context"
	"fmt"
)

const (
	defaultNextPageNum = 1
	defaultTotalPages  = 1
	defaultPageSize    = 100
)

type (
	Loader[T any] interface {
		Load(ctx context.Context, pageNum, pageSize uint) ([]T, error)
	}

	NewTotalLoader[T any] interface {
		Load(ctx context.Context, pageNum, pageSize uint) (page []T, newTotalCount uint, err error)
	}

	Pager[T any] struct {
		// elements count per page, must be positive.
		pageSize uint
		// total pages count.
		// If you don't know the total pages count, you do not set this value.
		// totalPagesCount will be updated after each call of NextWithNewTotal.
		totalPagesCount uint
		// next page number, that will be loaded in next call of Next or NextWithNewTotal.
		// 1, 2, 3, ..., totalPagesCount.
		nextPageNum uint
		// nextPageLoader or nextPageWithNewTotalLoader must be set.
		// If you know the total pages count, you should set nextPageLoader value.
		nextPageLoader Loader[T]
		// nextPageWithNewTotalLoader or nextPageLoader must be set.
		// If you do not know the total pages count, you should set nextPageWithNewTotalLoader value.
		nextPageWithNewTotalLoader NewTotalLoader[T]
	}
)

func New[T any](opts ...Option[T]) (*Pager[T], error) {
	pager := &Pager[T]{
		pageSize:                   defaultPageSize,
		totalPagesCount:            defaultTotalPages,
		nextPageNum:                defaultNextPageNum,
		nextPageLoader:             nil,
		nextPageWithNewTotalLoader: nil,
	}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}

	if pager.nextPageLoader == nil && pager.nextPageWithNewTotalLoader == nil {
		return nil, fmt.Errorf("next page loader is required")
	}
	if pager.nextPageNum > pager.totalPagesCount {
		return nil, fmt.Errorf("next page number must be less or equal to total pages count")
	}

	return pager, nil
}

// Next returns the next page of elements.
// Use this method if you set [nextPageLoader] value in New.
func (p *Pager[T]) Next(ctx context.Context) ([]T, error) {
	isAllPagesAlreadyLoaded := p.nextPageNum > p.totalPagesCount
	if isAllPagesAlreadyLoaded {
		return nil, nil
	}

	page, err := p.nextPageLoader.Load(ctx, p.nextPageNum, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("page %d: %w", p.nextPageNum, err)
	}
	p.nextPageNum++
	return page, nil
}

// All returns all elements from all pages.
// It uses Next method to load all pages.
// Use this method if you set [nextPageLoader] value in New.
func (p *Pager[T]) All(ctx context.Context) ([]T, error) {
	var allPages []T
	for {
		page, err := p.Next(ctx)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		allPages = append(allPages, page...)
	}
	return allPages, nil
}

// NextWithNewTotal returns the next page of elements.
// Use this method if you set [nextPageWithNewTotalLoader] value in New.
func (p *Pager[T]) NextWithNewTotal(ctx context.Context) ([]T, error) {
	isAllPagesAlreadyLoaded := p.nextPageNum > p.totalPagesCount
	if isAllPagesAlreadyLoaded {
		return nil, nil
	}

	page, newTotal, err := p.nextPageWithNewTotalLoader.Load(ctx, p.nextPageNum, p.pageSize)
	if err != nil {
		return nil, fmt.Errorf("page %d: %w", p.nextPageNum, err)
	}
	p.nextPageNum++
	p.totalPagesCount = TotalCountToTotalPagesCount(newTotal, p.pageSize)
	return page, nil
}

// AllWithNewTotal returns all elements from all pages.
// It uses NextWithNewTotal method to load all pages.
// Use this method if you set [nextPageWithNewTotalLoader] value in New.
func (p *Pager[T]) AllWithNewTotal(ctx context.Context) ([]T, error) {
	var allPages []T
	for {
		page, err := p.NextWithNewTotal(ctx)
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}
		allPages = append(allPages, page...)
	}
	return allPages, nil
}

func TotalCountToTotalPagesCount(totalCount, pageSize uint) uint {
	if pageSize == 0 {
		return 0
	}
	return (totalCount + pageSize - 1) / pageSize
}
