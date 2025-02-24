package utils

import (
	"sync"
	"time"
)

type AsyncBatchProcessor[T any] interface {
	Add(v T)
	Trigger()
	Stop()
}

type asyncBatchProcessor[T any] struct {
	values    chan T         // Buffered channel for values
	read      chan struct{}  // Read values channel
	stop      chan struct{}  // Stop signal
	wg        sync.WaitGroup // WaitGroup for cleanup
	processFn func([]T)      // Function to process batches
}

// NewAsyncBatchProcessor initializes a new AsyncBatchProcessor
func NewAsyncBatchProcessor[T any](interval time.Duration, bufferSize int, processFn func([]T)) AsyncBatchProcessor[T] {
	dr := &asyncBatchProcessor[T]{
		values:    make(chan T, bufferSize),
		read:      make(chan struct{}, 1), // Buffered to prevent blocking
		stop:      make(chan struct{}),
		processFn: processFn,
	}

	// Start the ticker goroutine
	ticker := time.NewTicker(interval)
	dr.wg.Add(1)
	go func() {
		defer dr.wg.Done()
		for {
			select {
			case <-ticker.C: // Time-based trigger
				dr.triggerRead()
			case <-dr.stop: // Stop the reader
				ticker.Stop()
				return
			}
		}
	}()

	// Start the batch processing goroutine
	dr.wg.Add(1)
	go dr.run()

	return dr
}

// run listens for trigger events and processes batches
func (dr *asyncBatchProcessor[T]) run() {
	defer dr.wg.Done()

	for {
		select {
		case <-dr.read:
			dr.processBatch()
		case <-dr.stop:
			dr.processBatch() // Final process before stopping
			return
		}
	}
}

// triggerRead sends a non-blocking signal to process the batch
func (dr *asyncBatchProcessor[T]) triggerRead() {
	select {
	case dr.read <- struct{}{}:
	default: // Prevent blocking if already triggered
	}
}

// processBatch reads all available items and processes them
func (dr *asyncBatchProcessor[T]) processBatch() {
	var batch []T

	for {
		select {
		case v := <-dr.values:
			batch = append(batch, v)
		default:
			if len(batch) > 0 {
				dr.processFn(batch) // Process batch
			}

			return
		}
	}
}

// Add sends a value to the reader for delayed processing
func (dr *asyncBatchProcessor[T]) Add(value T) {
	dr.values <- value
}

// Trigger forces a processing cycle
func (dr *asyncBatchProcessor[T]) Trigger() {
	dr.triggerRead()
}

// Stop shuts down the reader gracefully
func (dr *asyncBatchProcessor[T]) Stop() {
	close(dr.stop)
	dr.wg.Wait()
}
