package subpub

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// MessageHandler is a callback function that processes messages delivered to subscribers.
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for.
	Unsubscribe()
}
type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given subject.
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg interface{}) error
	// Close will shutdown sub-pub system.
	// May be blocked by data delivery until the context is canceled.
	Close(ctx context.Context) error
}

type subscriptionPublisher struct {
	mu       sync.RWMutex
	handlers map[string][]*subscription
	wg       sync.WaitGroup
	closed   atomic.Bool
	counter  atomic.Int64
}

type subscription struct {
	subject string
	id      int64
	handler MessageHandler
	subPub  *subscriptionPublisher
}

func (s *subscription) Unsubscribe() {
	if val, ok := s.subPub.handlers[s.subject]; ok {
		list := val

		for i, handler := range list {
			if handler.id == s.id {
				copy(list[i:], list[i+1:])

				list[len(list)-1] = nil
				list = list[:len(list)-1]

				break
			}
		}

		s.subPub.handlers[s.subject] = list
	}
}

func (s *subscriptionPublisher) Subscribe(subject string, cb MessageHandler) (Subscription, error) {

	if cb == nil {
		return nil, errors.New("nil handler")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed.Load() {
		return nil, errors.New("SubPub closed")
	}

	sub := &subscription{
		subject: subject,
		id:      s.counter.Add(1),
		handler: cb,
		subPub:  s,
	}

	s.handlers[subject] = append(s.handlers[subject], sub)
	return sub, nil
}

func (s *subscriptionPublisher) Publish(subject string, msg interface{}) error {
	if s.closed.Load() {
		return errors.New("SubPub closed")
	}

	s.mu.RLock()

	handlers := make([]*subscription, len(s.handlers[subject]))

	copy(handlers, s.handlers[subject])
	s.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	s.wg.Add(len(handlers))

	for _, h := range handlers {
		go func(handler MessageHandler) {
			defer s.wg.Done()
			handler(msg)
		}(h.handler)
	}

	return nil
}

func (s *subscriptionPublisher) Close(ctx context.Context) error {

	if !s.closed.CompareAndSwap(false, true) {
		return errors.New("SubPub closed")
	}

	done := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.mu.Lock()
		s.handlers = make(map[string][]*subscription)
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		s.closed.Swap(false)
		return ctx.Err()
	}
}

func (s *subscriptionPublisher) runCallback(handler MessageHandler, msg interface{}) {
	defer s.wg.Done()
	handler(msg)
}

func NewSubPub() SubPub {
	return &subscriptionPublisher{
		handlers: make(map[string][]*subscription),
	}
}
