package etcd

import (
	"context"
	"xdp-banner/pkg/log"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

// ListerWatcher is any object that knows how to perform an initial list and start a watch on a resource.
type ListerWatcher interface {
	Lister
	Watcher
}

// Lister is any object that knows how to perform an initial list.
type Lister interface {
	// List should return a list type object; the Items field will be extracted, and the
	// ResourceVersion field will be used to start the watch in the right place.
	List(ctx context.Context, opt ListOption) (PagedList, error)
}

// Watcher is any object that knows how to start a watch on a resource.
type Watcher interface {
	// Watch should begin a watch at the specified version.
	//
	// If Watch returns an error, it should handle its own cleanup, including
	// but not limited to calling Stop() on the watch, if one was constructed.
	// This allows the caller to ignore the watch, if the error is non-nil.
	Watch(opt WatchOption) (WatchController, error)
}

// WatchController can be implemented by anything that knows how to watch and report changes.
type WatchController interface {
	// Stop tells the producer that the consumer is done watching, so the
	// producer should stop sending events and close the result channel. The
	// consumer should keep watching for events until the result channel is
	// closed.
	//
	// Because some implementations may create channels when constructed, Stop
	// must always be called, even if the consumer has not yet called
	// ResultChan().
	//
	// Only the consumer should call Stop(), not the producer. If the producer
	// errors and needs to stop the watch prematurely, it should instead send
	// an error event and close the result channel.
	Stop()

	// ResultChan returns a channel which will receive events from the event
	// producer. If an error occurs or Stop() is called, the producer must
	// close this channel and release any resources used by the watch.
	// Closing the result channel tells the consumer that no more events will be
	// sent.
	ResultChan() <-chan Event
}

type PagedList struct {
	TotalCount  int64 // TotalCount all keys count
	TotalPage   int64 // TotalPage all pages count
	CurrentPage int64 // CurrentPage current page
	NextCursor  Key   // NextCursor next page start key, you should use it in next ListWithPagination call
	Revision    int64 //  Revision id of the response
	Items       *Items
}

func (p *PagedList) More() bool {
	return p.NextCursor != ""
}

type Items struct {
	Keys   []Key // CurrentPage keys
	Values []any // CurrentPage values
}

func NewItems(cap int) *Items {
	return &Items{
		Keys:   make([]Key, 0, cap),
		Values: make([]any, 0, cap),
	}
}

// Append append a key-value pair
func (i *Items) Append(key Key, value any) {
	i.Keys = append(i.Keys, key)
	i.Values = append(i.Values, value)
}

func (i *Items) AppendAll(i2 *Items) {
	i.Keys = append(i.Keys, i2.Keys...)
	i.Values = append(i.Values, i2.Values...)
}

// DropHead drop head x items
func (i *Items) DropHead(x int) {
	i.Keys = i.Keys[x:]
	i.Values = i.Values[x:]
}

func (i Items) Iterator() func(yield func(Key, any) bool) {
	return func(yield func(Key, any) bool) {
		for index, key := range i.Keys {
			if !yield(key, i.Values[index]) {
				return
			}
		}
	}
}

func (i Items) Len() int {
	return len(i.Keys)
}

// ItemConvert allow caller convert etcd raw response to a custom key-value pair
type ItemConvert = func(key Key, value *mvccpb.KeyValue) (Key, any)

type ListOption struct {
	Prefix   Key
	Size     int64
	Cursor   Key
	Revision int64
	Convert  ItemConvert
}

// ListWithPagination lists keys with pagination.
func (c *client) List(ctx context.Context, opt ListOption) (PagedList, error) {
	if opt.Size <= 0 {
		return PagedList{}, ErrInvalidPageSize
	}

	// set default convert
	if opt.Convert == nil {
		opt.Convert = func(key Key, value *mvccpb.KeyValue) (Key, any) {
			return key, string(value.Value)
		}
	}

	initList := opt.Cursor == "" // first list call

	// set key range
	startKey := opt.Cursor
	if initList {
		startKey = opt.Prefix
	}
	endKey := clientv3.GetPrefixRangeEnd(opt.Prefix)

	limit := int64(opt.Size)
	if !initList {
		// we should discard the first key, since the first key is a cursor
		// it's returned by the former ListWithPagination call
		limit++
	}

	txn, cancel := c.Txn(ctx)
	defer cancel()

	getOpt := []clientv3.OpOption{
		clientv3.WithRange(endKey),
		clientv3.WithLimit(limit),
	}

	if opt.Revision != 0 {
		getOpt = append(getOpt, clientv3.WithRev(opt.Revision))
	}

	resp, err := txn.Then(
		clientv3.OpGet(startKey, getOpt...),
		// only for get the total count
		clientv3.OpGet(opt.Prefix, clientv3.WithPrefix(), clientv3.WithCountOnly()),
	).Commit()
	if err != nil {
		return PagedList{}, err
	}

	kvResp := resp.Responses[0].GetResponseRange()
	countResp := resp.Responses[1].GetResponseRange()

	if countResp.Count == 0 {
		return PagedList{
			TotalCount:  0,
			TotalPage:   0,
			CurrentPage: 0,
			NextCursor:  "",
			Items:       &Items{},
			Revision:    resp.Header.Revision,
		}, nil
	}

	// kv
	kvs := kvResp.Kvs
	nextCursor := ""
	var items *Items

	if len(kvs) > 0 {
		items = NewItems(len(kvs))

		for _, kv := range kvs {
			k, v := opt.Convert(Key(kv.Key), kv)
			items.Append(k, v)
		}
		// discard the first key
		if !initList {
			items.DropHead(1)
		}

		// more page
		if kvResp.More {
			nextCursor = items.Keys[len(items.Keys)-1]
		}
	}

	totalCount := countResp.Count
	totalPage := ceil(totalCount, int64(opt.Size))
	var currentPage int64
	if initList {
		currentPage = 1
	} else {
		if kvResp.Count == 0 {
			currentPage = 0
		} else {
			currentPage = ceil(countResp.Count-kvResp.Count+2, int64(opt.Size))
		}
	}

	return PagedList{
		TotalCount:  totalCount,
		TotalPage:   totalPage,
		CurrentPage: currentPage,
		Items:       items,
		NextCursor:  nextCursor,
	}, nil
}

// ceil int version of math.Ceil to avoid float calculation
func ceil(x, y int64) int64 {
	return (x + y - 1) / y
}

type WatchOption struct {
	Prefix   Key
	Revision int64
	Convert  ItemConvert
}

func (cli *client) Watch(opt WatchOption) (WatchController, error) {
	ctx, cancel := context.WithCancel(context.Background())

	watchOpts := []clientv3.OpOption{
		clientv3.WithPrefix(),
	}
	if opt.Revision != 0 {
		watchOpts = append(watchOpts, clientv3.WithRev(opt.Revision))
	}

	watchChan := cli.Client.Watch(ctx, opt.Prefix, watchOpts...)
	c := newWatchController(ctx, cancel, watchChan, opt.Convert)
	go c.run()

	return c, nil
}

type watchControllerImpl struct {
	ctx        context.Context
	cancel     context.CancelFunc
	watchChan  clientv3.WatchChan
	resultChan chan Event
	convert    ItemConvert
}

func newWatchController(ctx context.Context, cancel context.CancelFunc, watchChan clientv3.WatchChan, convert ItemConvert) *watchControllerImpl {
	if convert == nil {
		// set default convert
		convert = func(key Key, value *mvccpb.KeyValue) (Key, any) {
			return key, string(value.Value)
		}
	}

	return &watchControllerImpl{
		ctx:        ctx,
		cancel:     cancel,
		watchChan:  watchChan,
		resultChan: make(chan Event),
		convert:    convert,
	}
}

func (w *watchControllerImpl) run() {
	for {
		select {
		case <-w.ctx.Done():
			close(w.resultChan)
			return
		case watchResp := <-w.watchChan:
			revision := watchResp.Header.Revision
			if watchResp.Err() != nil {
				w.resultChan <- Event{
					Type:     Error,
					Value:    watchResp.Err(),
					Revision: revision,
				}
				continue
			}
			for _, event := range watchResp.Events {

				// check valid event type
				switch event.Type {
				case clientv3.EventTypePut, clientv3.EventTypeDelete:
				default:
					log.Warn("receive unexpected type from etcd", zap.String("type", event.Type.String()))
					continue
				}

				k, v := w.convert(Key(event.Kv.Key), event.Kv)

				w.resultChan <- Event{
					Type:     EventType(event.Type.String()),
					Key:      k,
					Value:    v,
					LeaseID:  event.Kv.Lease,
					Revision: revision,
				}
			}
		}
	}
}

func (w *watchControllerImpl) Stop() {
	w.cancel()
}

func (w *watchControllerImpl) ResultChan() <-chan Event {
	return w.resultChan
}

// Event represents a single event from a watch stream.
type Event struct {
	// Type is the type of event (Added, Modified, Deleted).
	Type EventType
	// Key is the key of the event.
	Key Key
	// Value is the value of the event.
	Value any
	// LeaseID is the lease ID of the event.
	LeaseID int64
	// Revision is the etcd revision of the event.
	Revision int64
}

type EventType string

const (
	Put    EventType = "PUT"
	Delete EventType = "DELETE"
	Error  EventType = "ERROR"
)
