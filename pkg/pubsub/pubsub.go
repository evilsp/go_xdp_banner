package pubsub

import "sync"

type PubSub struct {
	sync.Mutex
	subscribers map[string]chan *Message
}

type Message struct {
	Key string
	Val any
}

func New() *PubSub {
	return &PubSub{
		subscribers: make(map[string]chan *Message),
	}
}

func (ps *PubSub) Subscribe(key string) <-chan *Message {
	notify := make(chan *Message)

	ps.Lock()
	defer ps.Unlock()

	ps.subscribers[key] = notify

	return notify
}

func (ps *PubSub) Unsubscribe(key string) {
	ps.Lock()
	defer ps.Unlock()

	close(ps.subscribers[key])
	delete(ps.subscribers, key)
}

func (ps *PubSub) Publish(msg string, val any) {
	ps.Lock()
	defer ps.Unlock()

	for _, ch := range ps.subscribers {
		select {
		case ch <- &Message{Key: msg, Val: val}:
		default:
		}
	}
}
