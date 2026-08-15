package progress

import (
	"github.com/cheggaaa/pb/v3"
)

type Bar struct {
	bar  *pb.ProgressBar
	pool *Pool
}

// AddTotal adds to the total bar value
func (b *Bar) AddTotal(n int) {
	if b.bar == nil {
		return
	}
	b.bar.AddTotal(int64(n))
	b.pool.redraw()
}

// Increment atomically increments the progress
func (b *Bar) Increment() {
	if b.bar == nil {
		return
	}
	b.bar.Increment()
	b.pool.redraw()
}

// Finish stops the bar
func (b *Bar) Finish() {
	if b.bar == nil {
		return
	}
	b.bar.Finish()
	b.pool.redraw()
}

// Add adding given int value to bar value
func (b *Bar) Add(value int) {
	if b.bar == nil {
		return
	}
	b.bar.Add(value)
	b.pool.redraw()
}
