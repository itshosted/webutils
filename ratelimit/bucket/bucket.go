package bucket
// Limit the amount of requests per second.
// Based off: http://en.wikipedia.org/wiki/Leaky_bucket
import (
	"math"
	"time"
)

type Bucket struct {
	Fillrate   float64
	Capacity   float64
	Available  float64
	LastUpdate time.Time

	Delay        time.Duration
	DelayUntil   time.Time
	DelayCounter int
}

// Increase request counter by amount.
// Return false if limit is reached and show available in second arg
func (b *Bucket) Request(amount float64) (bool, float64) {
	now := time.Now()

	/* Are we delaying requests? */
	if b.DelayUntil.Unix() > now.Unix() {
		b.DelayCounter++
		b.DelayUntil = time.Now().Add(b.Delay)
		return false, b.Available
	}

	timeDiff := now.Sub(b.LastUpdate).Seconds()
	b.Available = math.Min(b.Capacity, b.Available+(timeDiff*b.Fillrate))
	b.LastUpdate = now

	if b.Available >= amount {
		b.Available -= amount
		return true, b.Available
	} else {
		b.DelayCounter = 1
		b.DelayUntil = time.Now().Add(b.Delay)
		return false, b.Available
	}
}

// Create new bucket.
//
// fillrate = Amount of requests per second
// capacity = Extra requests allowed a-top fillrate
// delay = Time delay request if ratelimited
//
// Example: fillrate=10 capacity=10
//  this allows 10reqs/sec and if surpassed allow 10 reqs more
//  before returning false with Request()
func New(fillrate float64, capacity float64, delay time.Duration) *Bucket {
	return &Bucket{
		Fillrate:   fillrate,
		Capacity:   capacity,
		Available:  math.Max(fillrate, capacity),
		LastUpdate: time.Now(),
		Delay:      delay,
	}
}
