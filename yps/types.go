package yps

import "time"

type Page struct {
	Content      string
	GoogleFormID string
	Updated      time.Time
}

func (p *Page) ToHTML() string {
	//TODO(dan): return google form too
	return p.Content
}
