package bucket

import (
	"fmt"
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

func (b *Bucket) Request(amount float64) (bool, float64) {
	/* Get current TS */
	now := time.Now()

	/* Are we delaying requests? */
	if b.DelayUntil.Unix() > now.Unix() {
		b.DelayCounter++
		return false, b.Available
	}

	/* Get elapsed time */
	timeDiff := now.Sub(b.LastUpdate).Seconds()

	/* Calculate bucket fill based on last request */
	b.Available = math.Min(b.Capacity, b.Available+(timeDiff*b.Fillrate))

	/* Update TS */
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

func New(fillrate float64, capacity float64, delay time.Duration) *Bucket {
	return &Bucket{
		Fillrate:   fillrate,
		Capacity:   capacity,
		Available:  capacity,
		LastUpdate: time.Now(),
		Delay:      delay,
	}
}
