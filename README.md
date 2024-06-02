# Number pager

Go-Num-Pager is a paginator library for Go.
It helps you to paginate your data in a simple way.
This library uses an integer number as the key of a page.

If you want to paginate your data with a **string** key, you can use 
[go-string-pager](https://github.com/Nizom98/go-string-pager).

## Installation
```bash
go get github.com/Nizom98/go-num-pager
```

## Usage
If you know the total number of pages.

```go
package main

type MyLoader struct{}

func main() {
	myLoader := &MyLoader{}

	pager, _ := page.New[int](
		page.WithTotalCount[int](100),
		page.WithNextPageLoader[int](myLoader),
	)

	result, _ := pager.All(context.Background())
	fmt.Println(result)
}

func (l *MyLoader) Load(_ context.Context, pageNum, pageSize int) ([]int, error) {
	requstBody := map[string]interface{}{
		"limit":    pageSize,
		"start_at": pageNum,
	}
	// TODO: write your own logic to load page result
	return pageResult, nil
}
```

If you don't know the total number of pages.

```go
package main

type MyLoader struct{}

func main() {
	myLoader := &MyLoader{}

	pager, _ := page.New[int](
		page.WithNextPageLoaderWithNewTotal[int](myLoader),
	)

	result, _ := pager.All(context.Background())
	fmt.Println(result)
}

func (l *MyLoader) Load(
	_ context.Context,
	pageNum int,
	pageSize int,
) (pageResult []int, totalCount int, err error) {
	body := map[string]interface{}{
		"limit":    pageSize,
		"start_at": pageNum,
	}
	// TODO: write your own logic to load page result and new total count
	return pageResult, totalCount, nil
}
```