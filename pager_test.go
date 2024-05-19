package page

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
				totalCount:      defaultTotalCount,
				nextPageStartAt: defaultNextPageStartAt,
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
				totalCount:                      defaultTotalCount,
				nextPageStartAt:                 defaultNextPageStartAt,
				nextPageLoaderWithNewTotalCount: newTotalLoader,
			},
			wantErr: require.NoError,
		},
		{
			name: "custom options",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithNextPageStartAt[int](9),
				WithTotalCount[int](99),
				WithPageSize[int](999),
			},
			want: &Pager[int]{
				pageSize:                        999,
				totalCount:                      99,
				nextPageStartAt:                 9,
				nextPageLoaderWithNewTotalCount: newTotalLoader,
			},
			wantErr: require.NoError,
		},
		{
			name: "zero nextPageStartAt",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithNextPageStartAt[int](0),
				WithTotalCount[int](99),
				WithPageSize[int](999),
			},
			want: &Pager[int]{
				pageSize:                        999,
				totalCount:                      99,
				nextPageStartAt:                 0,
				nextPageLoaderWithNewTotalCount: newTotalLoader,
			},
			wantErr: require.NoError,
		},
		{
			name: "invalid nextPageStartAt",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithNextPageStartAt[int](-1),
				WithTotalCount[int](99),
				WithPageSize[int](999),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "next page start position must not be negative")
			},
		},
		{
			name: "invalid total pages count",
			opts: []Option[int]{
				WithNextPageLoaderWithNewTotal[int](newTotalLoader),
				WithTotalCount[int](0),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "total count must be positive")
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
				WithNextPageStartAt[int](2),
				WithTotalCount[int](1),
			},
			want: nil,
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "next page start position must be less than total count")
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
	defer ctrl.Finish()

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
					Load(gomock.Any(), defaultNextPageStartAt, defaultPageSize).
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
				loader.EXPECT().
					Load(gomock.Any(), defaultNextPageStartAt, defaultPageSize).
					Times(1).
					Return([]int{}, nil)
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
					Load(gomock.Any(), defaultNextPageStartAt, defaultPageSize).
					Times(1).
					Return(nil, errors.New("test error"))
				return loader
			},
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page start at 0: test error")
			},
		},
		{
			name: "start from 1",
			loader: func() Loader[int] {
				loader := NewMockLoader[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 1, defaultPageSize).
					Times(1).
					Return([]int{3, 4}, nil)
				return loader
			},
			opts: []Option[int]{
				WithNextPageStartAt[int](1),
				WithTotalCount[int](2),
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
	defer ctrl.Finish()

	cases := []struct {
		name           string
		loader         func() LoaderWithNewTotal[int]
		opts           []Option[int]
		want           []int
		wantTotalCount int
		wantErr        require.ErrorAssertionFunc
	}{
		{
			name: "all ok, default options",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 0, 100).
					Times(1).
					Return([]int{1, 2, 3}, 3, nil)
				return loader
			},
			want:           []int{1, 2, 3},
			wantTotalCount: 3,
			wantErr:        require.NoError,
		},
		{
			name: "got 0 total count",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 0, 100).
					Times(1).
					Return([]int{1, 2, 3}, 0, nil)
				return loader
			},
			want:           []int{1, 2, 3},
			wantTotalCount: 0,
			wantErr:        require.NoError,
		},
		{
			name: "loader returned empty page",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 0, 100).
					Times(1).
					Return([]int{}, 101, nil)
				return loader
			},
			want:           []int{},
			wantTotalCount: 101,
			wantErr:        require.NoError,
		},
		{
			name: "loader returned an error",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 0, 100).
					Times(1).
					Return(nil, 0, errors.New("test error"))
				return loader
			},
			wantTotalCount: 1, // should not be updated
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page start at 0: test error")
			},
		},
		{
			name: "start from 0",
			loader: func() LoaderWithNewTotal[int] {
				loader := NewMockLoaderWithNewTotal[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), 0, 100).
					Times(1).
					Return([]int{5, 6}, 101, nil)
				return loader
			},
			opts:           nil,
			want:           []int{5, 6},
			wantTotalCount: 101,
			wantErr:        require.NoError,
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
			require.Equal(t, tc.wantTotalCount, pager.totalCount)
		})
	}
}

func TestPager_All(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		name    string
		loader  func() Loader[int]
		opts    []Option[int]
		want    []int
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "get 3 pages",
			loader: func() Loader[int] {
				loader := &fakeLoader{
					pageByStartAt: map[int][]int{
						0: {1, 2, 3},
						3: {4, 5, 6},
						6: {7, 8},
					},
					errOnPageStartAt: -1, // no error
				}
				return loader
			},
			opts: []Option[int]{
				WithPageSize[int](3),
				WithTotalCount[int](8),
			},
			want:    []int{1, 2, 3, 4, 5, 6, 7, 8},
			wantErr: require.NoError,
		},
		{
			name: "get 2 pages, an error at the third page",
			loader: func() Loader[int] {
				loader := &fakeLoader{
					pageByStartAt: map[int][]int{
						0: {1, 2, 3},
						3: {4, 5, 6},
						6: {7, 8},
					},
					errOnPageStartAt: 6,
				}
				return loader
			},
			opts: []Option[int]{
				WithPageSize[int](3),
				WithTotalCount[int](8),
			},
			want: []int{1, 2, 3, 4, 5, 6},
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page start at 6: test error")
			},
		},
		{
			name: "get 1 page, an error at the second page",
			loader: func() Loader[int] {
				loader := &fakeLoader{
					pageByStartAt: map[int][]int{
						0: {1, 2, 3},
						3: {4, 5, 6},
						6: {7, 8},
					},
					errOnPageStartAt: 3,
				}
				return loader
			},
			opts: []Option[int]{
				WithPageSize[int](3),
				WithTotalCount[int](8),
			},
			want: []int{1, 2, 3},
			wantErr: func(t require.TestingT, err error, _ ...interface{}) {
				require.EqualError(t, err, "page start at 3: test error")
			},
		},
		{
			name: "nothing to load, must call loader 3 times",
			loader: func() Loader[int] {
				loader := NewMockLoader[int](ctrl)
				loader.EXPECT().
					Load(gomock.Any(), gomock.AnyOf(0, 3, 6), 3).
					Times(3).
					Return([]int{}, nil)
				return loader
			},
			opts: []Option[int]{
				WithPageSize[int](3),
				WithTotalCount[int](8),
			},
			want:    []int{},
			wantErr: require.NoError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.opts = append(tc.opts, WithNextPageLoader[int](tc.loader()))
			pager, err := New[int](tc.opts...)
			require.NoError(t, err)

			got, err := pager.All(context.Background())
			tc.wantErr(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

type fakeLoader struct {
	pageByStartAt    map[int][]int
	errOnPageStartAt int
}

func (l *fakeLoader) Load(_ context.Context, pageStartAt, _ int) (page []int, err error) {
	if l.errOnPageStartAt == pageStartAt {
		return nil, errors.New("test error")
	}
	return l.pageByStartAt[pageStartAt], nil
}
