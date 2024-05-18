package page

import (
	"context"
	"errors"
	"go.uber.org/mock/gomock"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTotalCountToTotalPagesCount(t *testing.T) {
	cases := []struct {
		name       string
		totalCount int
		pageSize   int
		want       int
	}{
		{
			name:       "zero total count and zero page size",
			totalCount: 0,
			pageSize:   0,
			want:       0,
		},
		{
			name:       "100 total count and 100 page size",
			totalCount: 100,
			pageSize:   100,
			want:       1,
		},
		{
			name:       "100 total count and 99 page size",
			totalCount: 100,
			pageSize:   99,
			want:       2,
		},
		{
			name:       "99 total count and 100 page size",
			totalCount: 99,
			pageSize:   100,
			want:       1,
		},
		{
			name:       "1 total count and 100 page size",
			totalCount: 1,
			pageSize:   100,
			want:       1,
		},
		{
			name:       "100 total count and 1 page size",
			totalCount: 100,
			pageSize:   1,
			want:       100,
		},
		{
			name:       "100 total count and 2 page size",
			totalCount: 100,
			pageSize:   2,
			want:       50,
		},
		{
			name:       "100 total count and 3 page size",
			totalCount: 100,
			pageSize:   3,
			want:       34,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t,
				tc.want,
				TotalCountToTotalPagesCount(tc.totalCount, tc.pageSize),
			)
		})
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	loader := NewMockLoader[int](ctrl)
	newTotalLoader := NewMockLoaderWithNewTotal[int](ctrl)

	cases := []struct {
		name    string
		opts    []Option[int]
		want    *Pager[int]
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "only with nextPageLoader option",
			opts: []Option[int]{
				WithNextPageLoader[int](loader),
			},
			want: &Pager[int]{
				pageSize:        defaultPageSize,
				totalPagesCount: defaultTotalPagesCount,
				nextPageNum:     defaultNextPageNum,
				nextPageLoader:  loader,
			},
			wantErr: require.NoError,
		},
		{
			name: "only with nextPageLoaderWithNewTotalCount option",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
			},
			want: &Pager[int]{
				pageSize:                        defaultPageSize,
				totalPagesCount:                 defaultTotalPagesCount,
				nextPageNum:                     defaultNextPageNum,
				nextPageLoaderWithNewTotalCount: newTotalLoader,
			},
			wantErr: require.NoError,
		},
		{
			name: "custom options",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithNexPageNum[int](9),
				WithTotalPagesCount[int](99),
				WithPageSize[int](999),
			},
			want: &Pager[int]{
				pageSize:                        999,
				totalPagesCount:                 99,
				nextPageNum:                     9,
				nextPageLoaderWithNewTotalCount: newTotalLoader,
			},
			wantErr: require.NoError,
		},
		{
			name: "invalid next page number",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithNexPageNum[int](0),
				WithTotalPagesCount[int](99),
				WithPageSize[int](999),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "next page number must be positive")
			},
		},
		{
			name: "invalid total pages count",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithTotalPagesCount[int](0),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "total pages count must be positive")
			},
		},
		{
			name: "invalid page size",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithPageSize[int](0),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page size must be positive")
			},
		},
		{
			name: "there is no next page loader",
			opts: nil,
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "next page loader is required")
			},
		},

		{
			name: "next page number is greater than total pages count",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithNexPageNum[int](2),
				WithTotalPagesCount[int](1),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "next page number must be less or equal to total pages count")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pager, err := New[int](tc.opts...)
			require.Equal(t, tc.want, pager)
			tc.wantErr(t, err)
		})
	}
}

func TestPager_Next_Loader(t *testing.T) {
	ctrl := gomock.NewController(t)

	cases := []struct {
		name    string
		loader  func() Loader[int]
		opts    []Option[int]
		want    []int
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "all ok, default options",
			loader: func() Loader[int] {
				loader := NewMockLoader[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), defaultNextPageNum, defaultPageSize).
					Times(1).
					Return([]int{1, 2, 3}, nil)
				return loader
			},
			want:    []int{1, 2, 3},
			wantErr: require.NoError,
		},
		{
			name: "loader returned empty page",
			loader: func() Loader[int] {
				loader := NewMockLoader[int](ctrl)
				loader.EXPECT().Load(gomock.Any(), defaultNextPageNum, defaultPageSize).Return([]int{}, nil)
				return loader
			},
			want:    []int{},
			wantErr: require.NoError,
		},
		{
			name: "loader returned an error",
			loader: func() Loader[int] {
				loader := NewMockLoader[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), defaultNextPageNum, defaultPageSize).
					Times(1).
					Return(nil, errors.New("test error"))
				return loader
			},
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page 1: test error")
			},
		},
		{
			name: "start from the second page",
			loader: func() Loader[int] {
				loader := NewMockLoader[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 2, defaultPageSize).
					Times(1).
					Return([]int{3, 4}, nil)
				return loader
			},
			opts: []Option[int]{
				WithNexPageNum[int](2),
				WithTotalPagesCount[int](2),
			},
			want:    []int{3, 4},
			wantErr: require.NoError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts = append(tc.opts, WithNextPageLoader[int](tc.loader()))
			pager, err := New[int](tc.opts...)
			require.NoError(t, err)

			got, err := pager.Next(context.Background())
			tc.wantErr(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestPager_Next_LoaderWithNewTotal(t *testing.T) {
	ctrl := gomock.NewController(t)

	cases := []struct {
		name                string
		loader              func() LoaderWithNewTotal[int]
		opts                []Option[int]
		want                []int
		wantTotalPagesCount int
		wantErr             require.ErrorAssertionFunc
	}{
		{
			name: "all ok, default options",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 1, 100).
					Times(1).
					Return([]int{1, 2, 3}, 1, nil)
				return loader
			},
			want:                []int{1, 2, 3},
			wantTotalPagesCount: 1,
			wantErr:             require.NoError,
		},
		{
			name: "got 0 total count",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 1, 100).
					Times(1).
					Return([]int{1, 2, 3}, 0, nil)
				return loader
			},
			want:                []int{1, 2, 3},
			wantTotalPagesCount: 0,
			wantErr:             require.NoError,
		},
		{
			name: "loader returned empty page",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 1, 100).
					Times(1).
					Return([]int{}, 101, nil)
				return loader
			},
			want:                []int{},
			wantTotalPagesCount: 2,
			wantErr:             require.NoError,
		},
		{
			name: "loader returned an error",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 1, 100).
					Times(1).
					Return(nil, 0, errors.New("test error"))
				return loader
			},
			wantTotalPagesCount: 1, // should not be updated
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page 1: test error")
			},
		},
		{
			name: "start from the second page",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 2, 100).
					Times(1).
					Return([]int{5, 6}, 101, nil)
				return loader
			},
			opts: []Option[int]{
				WithNexPageNum[int](2),
				WithTotalPagesCount[int](2),
			},
			want:                []int{5, 6},
			wantTotalPagesCount: 2,
			wantErr:             require.NoError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts = append(tc.opts, WithNextPageLoaderWithNewTotal[int](tc.loader()))
			pager, err := New[int](tc.opts...)
			require.NoError(t, err)

			got, err := pager.Next(context.Background())
			tc.wantErr(t, err)
			require.Equal(t, tc.want, got)
			require.Equal(t, tc.wantTotalPagesCount, pager.totalPagesCount)
		})
	}
}
