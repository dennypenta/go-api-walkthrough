package domain

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidUsername = errors.New("invalid username")
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func (u User) Validate() error {
	if len(u.Username) < 3 {
		return ErrInvalidUsername
	}
	return nil
}

type UserFilter struct {
	Limit  int
	Offset int
}

type PaginatedUserList struct {
	Users []User `json:"users"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`

	Page  int    `json:"page"`
	Pages int    `json:"pages"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

func (l *PaginatedUserList) EnrichHttpQueryLinks() {
	page := float64(l.Offset)/float64(l.Limit) + 1
	page = math.Round(page)
	l.Page = int(page)

	var pages float64 = 0
	if l.Offset != 0 {
		pages = (float64(l.Offset)) / float64(l.Limit)
		pages = math.Round(pages)
	}
	pages += float64(l.Total-l.Offset) / float64(l.Limit)
	pages = math.Round(pages)
	l.Pages = int(pages)

	if l.Offset != 0 {
		offset := l.Offset - l.Limit
		if offset < 0 {
			offset = 0
		}
		l.Prev = fmt.Sprintf("limit=%d&offset=%d", l.Limit, offset)
	}

	if l.Offset+l.Limit < l.Total {
		l.Next = fmt.Sprintf("limit=%d&offset=%d", l.Limit, l.Offset+l.Limit)
	}
}
