package etcd

import (
	"context"
)

const defaultPageSize = 500

type ListPageFunc func(ctx context.Context, opt ListOption) (PagedList, error)

// ListPager assists client code in breaking large list queries into multiple
// smaller chunks of PageSize or smaller. PageFn is expected to accept a
// metav1.ListOptions that supports paging and return a list. The pager does
// not alter the field or label selectors on the initial options list.
type ListPager struct {
	PageSize int64
	PageFn   ListPageFunc
}

// New creates a new pager from the provided pager function using the default
// options. It will fall back to a full list if an expiration error is encountered
// as a last resort.
func NewListPager(fn ListPageFunc) *ListPager {
	return &ListPager{
		PageSize: defaultPageSize,
		PageFn:   fn,
	}
}

// List returns a single list object, but attempts to retrieve smaller chunks from the
// server to reduce the impact on the server. If the chunk attempt fails, it will load
// the full list instead. The Limit field on options, if unset, will default to the page size.
//
// If items in the returned list are retained for different durations, and you want to avoid
// retaining the whole slice returned by p.PageFn as long as any item is referenced,
// use ListWithAlloc instead.
func (p *ListPager) List(ctx context.Context, options ListOption) (PagedList, error) {
	if options.Size == 0 {
		options.Size = p.PageSize
	}

	items := NewItems(20)
	pl := PagedList{
		Items: items,
	}

	for {
		select {
		case <-ctx.Done():
			return pl, ctx.Err()
		default:
		}

		obj, err := p.PageFn(ctx, options)
		if err != nil {
			return pl, err
		}

		items.AppendAll(obj.Items)
		pl.TotalCount = obj.TotalCount
		pl.TotalPage = obj.TotalPage
		pl.CurrentPage = obj.CurrentPage
		pl.Revision = obj.Revision

		// if we have no more items, return the list
		if !obj.More() {
			return pl, nil
		}

		// set the next loop up
		options.Cursor = obj.NextCursor
		options.Revision = obj.Revision
	}
}
