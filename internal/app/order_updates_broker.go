package app

import (
	"sync"

	"order-service/internal/domain"
)

type OrderUpdatesBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan *domain.Order]struct{}
}

func NewOrderUpdatesBroker() *OrderUpdatesBroker {
	return &OrderUpdatesBroker{
		subscribers: make(map[string]map[chan *domain.Order]struct{}),
	}
}

func (b *OrderUpdatesBroker) Subscribe(orderID string) (<-chan *domain.Order, func()) {
	ch := make(chan *domain.Order, 8)

	b.mu.Lock()
	if _, ok := b.subscribers[orderID]; !ok {
		b.subscribers[orderID] = make(map[chan *domain.Order]struct{})
	}
	b.subscribers[orderID][ch] = struct{}{}
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		subs, ok := b.subscribers[orderID]
		if !ok {
			return
		}

		if _, exists := subs[ch]; exists {
			delete(subs, ch)
			close(ch)
		}

		if len(subs) == 0 {
			delete(b.subscribers, orderID)
		}
	}

	return ch, cancel
}

func (b *OrderUpdatesBroker) Publish(order *domain.Order) {
	if order == nil {
		return
	}

	b.mu.RLock()
	subs := b.subscribers[order.ID]
	if len(subs) == 0 {
		b.mu.RUnlock()
		return
	}

	channels := make([]chan *domain.Order, 0, len(subs))
	for ch := range subs {
		channels = append(channels, ch)
	}
	b.mu.RUnlock()

	for _, ch := range channels {
		orderCopy := *order
		select {
		case ch <- &orderCopy:
		default:
		}
	}
}
