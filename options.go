package page

import "fmt"

type Option[T any] func(*Pager[T]) error

func WithNextPageStartAt[T any](startAt int) Option[T] {
	return func(p *Pager[T]) error {
		if startAt < 0 {
			return fmt.Errorf("next page start position must not be negative")
		}
		p.nextPageStartAt = startAt
		return nil
	}
}

func WithPageSize[T any](pageSize int) Option[T] {
	return func(p *Pager[T]) error {
		if pageSize <= 0 {
			return fmt.Errorf("page size must be positive")
		}
		p.pageSize = pageSize
		return nil
	}
}

func WithTotalCount[T any](totalCount int) Option[T] {
	return func(p *Pager[T]) error {
		if totalCount <= 0 {
			return fmt.Errorf("total count must be positive")
		}
		p.totalCount = totalCount
		return nil
	}
}

func WithNextPageLoader[T any](loader Loader[T]) Option[T] {
	return func(p *Pager[T]) error {
		if loader == nil {
			return fmt.Errorf("next page loader is required")
		}
		p.nextPageLoader = loader
		return nil
	}
}

func WithNextPageLoaderWithNewTotal[T any](loader LoaderWithNewTotal[T]) Option[T] {
	return func(p *Pager[T]) error {
		if loader == nil {
			return fmt.Errorf("next page loader with new total count is required")
		}
		p.nextPageLoaderWithNewTotalCount = loader
		return nil
	}
}
