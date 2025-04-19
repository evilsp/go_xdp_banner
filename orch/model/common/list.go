package common

type List struct {
	// TotalCount is the total number of configs.
	TotalCount int64 `json:"totalCount"`
	// TotalPage is the total number of pages with the given page size.
	TotalPage int64 `json:"totalPage"`
	// CurrentPage is the current page number.
	// The first page is 1. When request is invalid, CurrentPage is 0.
	CurrentPage int64 `json:"currentPage"`
	// HasNext is true if there is a next page.
	HasNext bool `json:"hasNext"`
	// NextCursor is the cursor for the next page. you should use it in next ListConfig call.
	NextCursor string `json:"nextCursor"`
}
