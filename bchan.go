package bchan

import (
	"sync"
)

// Bchan is an 1:N non-blocking value-loadable channel.
// The client needs to only know about one
// rule: after a receive on Ch, you must call Bchan.BcastAck().
//
type Bchan struct {
	Ch  chan interface{}
	mu  sync.Mutex
	on  bool
	cur interface{}
}

// New constructor should be told
// how many recipients are expected in
// expectedDiameter. If the expectedDiameter
// is wrong the Bchan will still function,
// but you may get slower concurrency
// than if the number is accurate. It
// is fine to overestimate the diameter;
// but the extra slots in the buffered channel
// take up some memory and need service time
// to be maintained.
func New(expectedDiameter int) *Bchan {
	return &Bchan{
		Ch: make(chan interface{}, expectedDiameter+1),
	}
}

// On turns on the broadcast channel without
// changing the value to be transmitted.
//
func (b *Bchan) On() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.on = true
	b.fill()
}

// Set stores a value to be broadcast
// and clears any prior queued up
// old values. Call On() after set
// to activate the new value.
// See also Bcast that does Set()
// followed by On() in one call.
//
func (b *Bchan) Set(val int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cur = val
	b.drain()
}

// Bcast is the common case of doing
// both Set() and then On() together
// to start broadcasting a new value.
//
func (b *Bchan) Bcast(val interface{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cur = val
	b.drain()
	b.on = true
	b.fill()
}

// Off turns off broadcasting
func (b *Bchan) Off() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.on = false
	b.drain()
}

// drain all messages, leaving Ch empty.
func (b *Bchan) drain() {
	// empty chan
	for {
		select {
		case <-b.Ch:
		default:
			return
		}
	}
}

// BcastAck is to be called immediately after
// the client receives on Ch. All
// clients on every receive must call BcastAck after receiving
// on the channel Ch. This makes such channels
// self-servicing, as BcastAck will re-fill the
// async channel with the current value.
func (b *Bchan) BcastAck() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.on {
		b.fill()
	}
}

// fill up the channel
func (b *Bchan) fill() {
	for {
		select {
		case b.Ch <- b.cur:
		default:
			return
		}
	}
}
