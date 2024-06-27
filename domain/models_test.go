package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginatedUserList_EnrichHttpQueryLinks(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		list PaginatedUserList

		expectedPrev  string
		expectedNext  string
		expectedPage  int
		expectedPages int
	}

	for _, tt := range []testCase{
		{
			name: "first page",
			list: PaginatedUserList{
				Users:  nil,
				Limit:  10,
				Offset: 0,
				Total:  20,
			},
			expectedPage:  1,
			expectedPages: 2,
			expectedPrev:  "",
			expectedNext:  "limit=10&offset=10",
		},
		{
			name: "first page, but no more content",
			list: PaginatedUserList{
				Users:  nil,
				Limit:  10,
				Offset: 0,
				Total:  10,
			},
			expectedPage:  1,
			expectedPages: 1,
			expectedPrev:  "",
			expectedNext:  "",
		},
		{
			name: "middle page",
			list: PaginatedUserList{
				Users:  nil,
				Limit:  10,
				Offset: 10,
				Total:  30,
			},
			expectedPrev:  "limit=10&offset=0",
			expectedNext:  "limit=10&offset=20",
			expectedPage:  2,
			expectedPages: 3,
		},
		{
			name: "last page",
			list: PaginatedUserList{
				Users:  nil,
				Limit:  10,
				Offset: 10,
				Total:  20,
			},
			expectedPrev:  "limit=10&offset=0",
			expectedNext:  "",
			expectedPage:  2,
			expectedPages: 2,
		},
		{
			name: "1.5 page",
			list: PaginatedUserList{
				Users:  nil,
				Limit:  10,
				Offset: 5,
				Total:  20,
			},
			expectedPrev:  "limit=10&offset=0",
			expectedNext:  "limit=10&offset=15",
			expectedPage:  2,
			expectedPages: 3,
		},
		{
			name: "2.5 page",
			list: PaginatedUserList{
				Users:  nil,
				Limit:  10,
				Offset: 15,
				Total:  20,
			},
			expectedPrev:  "limit=10&offset=5",
			expectedNext:  "",
			expectedPage:  3,
			expectedPages: 3,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.list.EnrichHttpQueryLinks()

			assert.Equal(t, tt.expectedPrev, tt.list.Prev)
			assert.Equal(t, tt.expectedNext, tt.list.Next)
			assert.Equal(t, tt.expectedPage, tt.list.Page)
			assert.Equal(t, tt.expectedPages, tt.list.Pages)
		})
	}
}
