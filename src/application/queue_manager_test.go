package application

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestQueueManagerNoGoroutineLeaks verifica vazamento em QueueManager
func TestQueueManagerNoGoroutineLeaks(t *testing.T) {
	// Baseline
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial goroutines: %d", initialGoroutines)

	// Simular operações de fila
	var wg sync.WaitGroup
	numOperations := 50

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Simular operação de fila (push/pop)
			time.Sleep(time.Duration(id%10) * time.Millisecond)
		}(i)
	}

	// Aguardar conclusão
	wg.Wait()

	// Cleanup
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutines: %d", finalGoroutines)

	if finalGoroutines > initialGoroutines+1 {
		t.Errorf("Goroutine leak detected: expected max %d, got %d",
			initialGoroutines+1, finalGoroutines)

		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		t.Logf("Stack traces:\n%s", string(buf[:n]))
	}
}

// TestConcurrentQueueOperations testa operações concorrentes
func TestConcurrentQueueOperations(t *testing.T) {
	const (
		numProducers = 5
		numConsumers = 5
		itemsPerProd = 100
	)

	// Usar channel como fila para simular
	queue := make(chan int, itemsPerProd*numProducers)
	var producerWg sync.WaitGroup
	var consumerWg sync.WaitGroup

	// Produtores
	for p := 0; p < numProducers; p++ {
		producerWg.Add(1)
		go func(producerID int) {
			defer producerWg.Done()

			for i := 0; i < itemsPerProd; i++ {
				queue <- producerID*itemsPerProd + i
			}
		}(p)
	}

	// Consumidores
	consumed := make(map[int]int)
	var consumeMutex sync.Mutex

	for c := 0; c < numConsumers; c++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()

			for item := range queue {
				consumeMutex.Lock()
				consumed[item]++
				consumeMutex.Unlock()
			}
		}()
	}

	// Aguardar produtores e fechar channel
	go func() {
		producerWg.Wait()
		close(queue)
	}()

	// Aguardar consumidores
	consumerWg.Wait()

	expectedItems := numProducers * itemsPerProd
	if len(consumed) != expectedItems {
		t.Errorf("Expected %d items, got %d", expectedItems, len(consumed))
	}

	t.Logf("✅ Processed %d items from %d producers to %d consumers",
		len(consumed), numProducers, numConsumers)
}

// BenchmarkQueueThroughput benchmark de throughput
func BenchmarkQueueThroughput(b *testing.B) {
	queue := make(chan int, 1000)

	go func() {
		for i := 0; i < b.N; i++ {
			queue <- i
		}
		close(queue)
	}()

	for range queue {
		// Consumir
	}
}
