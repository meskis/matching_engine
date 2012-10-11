package limitheap

import (
	"github.com/fmstephe/matching_engine/trade"
)

type limit struct {
	price int32
	head  *trade.Order
	tail  *trade.Order
	next  *limit
}

func newLimit(price int32, o *trade.Order) *limit {
	l := &limit{price: price, head: o, tail: o}
	o.Inward = &l.head
	return l
}

func (l *limit) push(o *trade.Order) {
	if l.head == nil {
		l.head = o
		o.Inward = &l.head
		l.tail = o
	} else {
		l.tail.Outward = o
		o.Inward = &l.tail.Outward
		l.tail = o
	}
}

func (l *limit) pop() *trade.Order {
	if l.head == nil {
		return nil
	}
	if l.head == l.tail {
		l.tail = nil
	}
	o := l.head
	l.head = o.Outward
	o.Inward = nil
	o.Outward = nil
	return o
}

func (l *limit) peek() *trade.Order {
	return l.head
}

func (l *limit) isEmpty() bool {
	return l.head == nil
}

func better(l1, l2 *limit, kind trade.OrderKind) bool {
	if kind == trade.BUY {
		return l2.price-l1.price < 0
	}
	return l1.price-l2.price < 0
}

type H struct {
	kind   trade.OrderKind
	limits *limitset
	orders *orderset
	heap   []*limit
	size   int
}

func New(kind trade.OrderKind, limitSetSize, ordersSize int32, heapSize int) *H {
	return &H{kind: kind, limits: newLimitSet(limitSetSize), orders: newOrderSet(ordersSize), heap: make([]*limit, 0, heapSize)}
}

func (h *H) Size() int {
	return h.size
}

func (h *H) Push(o *trade.Order) {
	h.orders.Put(o.Guid, o)
	lim := h.limits.Get(o.Price)
	if lim == nil {
		lim = newLimit(o.Price, o)
		h.limits.Put(o.Price, lim)
		h.heap = append(h.heap, lim)
		h.up(len(h.heap) - 1)
	} else {
		lim.push(o)
	}
	h.size++
}

func (h *H) Pop() *trade.Order {
	h.clearHead()
	if len(h.heap) == 0 {
		return nil
	}
	o := h.heap[0].pop()
	h.clearHead()
	h.orders.Remove(o.Guid)
	h.size--
	return o
}

func (h *H) Peek() *trade.Order {
	h.clearHead()
	if len(h.heap) == 0 {
		return nil
	}
	return h.heap[0].peek()
}

func (h *H) clearHead() {
	for len(h.heap) > 0 {
		lim := h.heap[0]
		if !lim.isEmpty() {
			return
		}
		n := len(h.heap) - 1
		h.heap[0] = h.heap[n]
		h.heap[n] = nil
		h.heap = h.heap[0:n]
		h.down(0)
		h.limits.Remove(lim.price)
	}
}

func (h *H) Remove(guid int64) *trade.Order {
	o := h.orders.Remove(guid)
	*o.Inward = o.Outward
	o.Inward = nil
	o.Outward = nil
	return o
}

func (h *H) Kind() trade.OrderKind {
	return h.kind
}

func (h *H) up(c int) {
	heap := h.heap
	for {
		p := (c - 1) / 2
		if p == c || better(heap[p], heap[c], h.kind) {
			break
		}
		heap[p], heap[c] = heap[c], heap[p]
		c = p
	}
}

func (h *H) down(p int) {
	n := len(h.heap)
	heap := h.heap
	for {
		c := 2*p + 1
		if c >= n {
			break
		}
		lc := c
		if rc := lc + 1; rc < n && !better(heap[lc], heap[rc], h.kind) {
			c = rc
		}
		if better(heap[p], heap[c], h.kind) {
			break
		}
		heap[p], heap[c] = heap[c], heap[p]
		p = c
	}
}
