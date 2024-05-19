# Number pager

Go-Num-Pager is a paginator library for Go.
It helps you to paginate your data in a simple way.


## Usage

If you know the total number of pages.
```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Nizom98/go-num-pager"
)

type MyLoader struct{}

func main() {
	myLoader := &MyLoader{}

	pager, err := page.New[int](
		page.WithPageSize[int](20),
		page.WithTotalCount[int](100),
		page.WithNextPageLoader[int](myLoader),
	)
	if err != nil {
		panic(err)
	}

	result, err := pager.All(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func (l *MyLoader) Load(_ context.Context, pageNum, pageSize int) ([]int, error) {
	body := map[string]interface{}{
		"limit":    pageSize,
		"start_at": pageNum,
	}

	bodyContent, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		"https://httpbin.org/post",
		"application/json",
		bytes.NewBuffer(bodyContent),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pageResult []int
	err = json.NewDecoder(resp.Body).Decode(&pageResult)
	if err != nil {
		return nil, err
	}

	return pageResult, nil
}
```
