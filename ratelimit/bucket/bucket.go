package bucket

import (
	"math"
	"time"
)

type Bucket struct {
	Fillrate   float64
	Capacity   float64
	Available  float64
	LastUpdate time.Time
}

func (b *Bucket) Request(amount float64) (bool, float64) {
	/* Get current TS */
	now := time.Now()

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
		return false, b.Available
	}
}

func New(fillrate float64, capacity float64) *Bucket {
	return &Bucket{
		Fillrate:   fillrate,
		Capacity:   capacity,
		Available:  capacity,
		LastUpdate: time.Now(),
	}
}
