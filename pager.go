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
		Load(ctx context.Context, page, pageSize uint) ([]T, error)
	}

	Pager[T any] struct {
		pageSize        uint
		totalPagesCount uint
		// 1, 2, 3, ..., totalPagesCount
		nextPageNum    uint
		nextPageLoader Loader[T]
	}
)

func New[T any](opts ...Option[T]) (*Pager[T], error) {
	pager := &Pager[T]{
		pageSize:        defaultPageSize,
		totalPagesCount: defaultTotalPages,
		nextPageNum:     defaultNextPageNum,
		// must be set by options
		nextPageLoader: nil,
	}
	for _, opt := range opts {
		if err := opt(pager); err != nil {
			return nil, err
		}
	}

	if pager.nextPageLoader == nil {
		return nil, fmt.Errorf("next page loader is required")
	}
	if pager.nextPageNum > pager.totalPagesCount {
		return nil, fmt.Errorf("next page number must be less or equal to total pages count")
	}

	return pager, nil
}

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

func TotalCountToTotalPagesCount(totalCount, pageSize uint) uint {
	if pageSize == 0 {
		return 0
	}
	return (totalCount + pageSize - 1) / pageSize
}
