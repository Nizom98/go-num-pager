package page

import "fmt"

type Option[T any] func(*Pager[T]) error

func WithNexPageNum[T any](page uint) Option[T] {
	return func(p *Pager[T]) error {
		if page <= 0 {
			return fmt.Errorf("next page number must be positive")
		}
		p.nextPageNum = page
		return nil
	}
}

func WithPageSize[T any](pageSize uint) Option[T] {
	return func(p *Pager[T]) error {
		if pageSize <= 0 {
			return fmt.Errorf("page size must be positive")
		}
		p.pageSize = pageSize
		return nil
	}
}

func WithTotalPages[T any](pagesCount uint) Option[T] {
	return func(p *Pager[T]) error {
		if pagesCount <= 0 {
			return fmt.Errorf("total pages count must be positive")
		}
		p.totalPagesCount = pagesCount
		return nil
	}
}
