package progress

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/NatoBoram/ipapm/env"
	"github.com/cheggaaa/pb/v3"
)

// Pool coordinates progress bars and log output sharing the same terminal.
//
// cheggaaa/pb's own Pool redraws bars on its own ticker goroutine, which races
// with anything else writing to the terminal (like our logger) and corrupts the
// display. ipfs/kubo avoids that by funnelling both the bar and its other
// output through a single sequenced writer that clears the bar line before
// printing and redraws it after; Pool does the same thing here, generalised to
// any number of bars, with its own ticker driving the redraw rate so bar
// updates from many goroutines don't each trigger a terminal write.
type Pool struct {
	enabled bool

	mu   sync.Mutex
	out  io.Writer
	bars []*pb.ProgressBar

	// drawn is the number of bar lines currently on screen
	drawn int
}

// NewPool initialises a pool. Bars are only drawn outside of production.
func NewPool(environment env.Environment) *Pool {
	return new(Pool{
		enabled: environment != env.Production,
		out:     os.Stdout,
	})
}

// Stop clears any remaining bars from the terminal.
func (p *Pool) Stop() error {
	if !p.enabled {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.clearLocked()

	return nil
}

// NewBar creates and registers a new bar with a prefix.
func (p *Pool) NewBar(name string) *Bar {
	if !p.enabled {
		return new(Bar{})
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	bar := pb.Simple.New(0).
		Set("prefix", name).
		Set(pb.Static, true).
		SetWriter(io.Discard)
	bar.Start()

	p.clearLocked()
	p.bars = append(p.bars, bar)
	p.redrawLocked()

	return &Bar{bar: bar, pool: p}
}

// Write implements [io.Writer] so the logger can route through the pool: the
// bars are cleared, the log line is printed above them, then they're redrawn.
func (p *Pool) Write(b []byte) (int, error) {
	if !p.enabled {
		return os.Stderr.Write(b)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.clearLocked()
	n, err := p.out.Write(b)
	p.redrawLocked()

	return n, err
}

func (p *Pool) redraw() {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.clearLocked()
	p.redrawLocked()
}

// clearLocked wipes the previously drawn bar lines. Caller must hold mu.
func (p *Pool) clearLocked() {
	if p.drawn == 0 {
		return
	}

	fmt.Fprintf(p.out, "\033[%dA", p.drawn)
	for range p.drawn {
		fmt.Fprint(p.out, "\033[2K\r\n")
	}
	fmt.Fprintf(p.out, "\033[%dA", p.drawn)
	p.drawn = 0
}

// redrawLocked drops finished bars and prints the remaining ones. Caller must
// hold mu.
func (p *Pool) redrawLocked() {
	var active []*pb.ProgressBar
	var inactive []*pb.ProgressBar
	for _, bar := range p.bars {
		if !bar.IsFinished() {
			active = append(active, bar)
		} else {
			inactive = append(inactive, bar)
		}
	}

	// Leave the completed bars on the screen, to be scrolled up by the next log
	// lines.
	for _, bar := range inactive {
		fmt.Fprintln(p.out, bar.String())
	}

	for _, bar := range active {
		fmt.Fprintln(p.out, bar.String())
	}

	p.bars = active
	p.drawn = len(active)
}
