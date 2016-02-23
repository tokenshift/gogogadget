package examples

import (
	"runtime"
	"sync"
	. "testing"
)

func init() {
	runtime.GOMAXPROCS(8)
}

func TestExampleCounterAgent(t *T) {
	counter := NewCounterAgent(0)
	defer counter.Close()

	var wg sync.WaitGroup

	for goroutine := 0; goroutine < 32; goroutine++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for action := 0; action < 10000; action++ {
				if action % 2 == 0 {
					counter.Add(1)
				} else {
					counter.Sub(1)
				}
			}
		}()
	}

	wg.Wait()

	if counter.Total() != 0 {
		t.Errorf("Expected 0, got %v", counter.Total())
	}
}

func TestExampleCounter(t *T) {
	counter := NewCounter(0)

	var wg sync.WaitGroup

	for goroutine := 0; goroutine < 32; goroutine++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for action := 0; action < 10000; action++ {
				if action % 2 == 0 {
					counter.Add(1)
				} else {
					counter.Sub(1)
				}
			}
		}()
	}

	wg.Wait()

	// Commented out because this is actually expected to fail.
	/*if counter.Total() != 0 {
		t.Errorf("Expected 0, got %v", counter.Total())
	}*/
}
