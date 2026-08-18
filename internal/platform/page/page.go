package page

import "windops/internal/fault"

type Request struct {
	Limit  int
	Offset int
	Sort   string
	Desc   bool
}
type Result[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (r Request) Normalize(allowed map[string]bool, fallback string) (Request, error) {
	if r.Limit == 0 {
		r.Limit = 50
	}
	if r.Limit < 1 || r.Limit > 200 {
		return Request{}, fault.New(fault.CodeInvalid, "page.normalize", "limit must be between 1 and 200")
	}
	if r.Offset < 0 {
		return Request{}, fault.New(fault.CodeInvalid, "page.normalize", "offset cannot be negative")
	}
	if r.Sort == "" {
		r.Sort = fallback
	}
	if !allowed[r.Sort] {
		return Request{}, fault.New(fault.CodeInvalid, "page.normalize", "unsupported sort field")
	}
	return r, nil
}
