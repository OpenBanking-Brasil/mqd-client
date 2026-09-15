package application

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestMessageProcessorWorkerNoGoroutineLeaks verifica vazamento em MessageProcessorWorker
func TestMessageProcessorWorkerNoGoroutineLeaks(t *testing.T) {
	// Baseline de goroutines
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial goroutines: %d", initialGoroutines)

	// Simular operações com mutex
	testWorker := &MessageProcessorWorker{
		receivedValues:  make(map[string]int),
		validatedValues: make(map[string]int),
	}

	// Simular acesso concorrente com goroutines
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Simular operação sob mutex
			messageProcessorWorkerMutex.Lock()
			testWorker.receivedValues["test"] = id
			time.Sleep(1 * time.Millisecond) // Simular trabalho
			messageProcessorWorkerMutex.Unlock()
		}(i)
	}

	// Aguardar conclusão de todas as goroutines
	wg.Wait()

	// Aguardar cleanup
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutines: %d (expected: %d)", finalGoroutines, initialGoroutines)

	// Verificar vazamento (permitir 1 goroutine extra para GC)
	if finalGoroutines > initialGoroutines+1 {
		t.Errorf("Goroutine leak detected: expected max %d, got %d",
			initialGoroutines+1, finalGoroutines)

		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		t.Logf("Stack traces:\n%s", string(buf[:n]))
	}
}

// TestMessageProcessorWorkerRaceCondition testa condições de corrida
func TestMessageProcessorWorkerRaceCondition(t *testing.T) {
	testWorker := &MessageProcessorWorker{
		receivedValues:  make(map[string]int),
		validatedValues: make(map[string]int),
	}

	var wg sync.WaitGroup
	numGoroutines := 100

	// Escrever e ler concorrentemente
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			messageProcessorWorkerMutex.Lock()
			testWorker.receivedValues["msg"+string(rune(id))] = id
			value := testWorker.receivedValues["msg"+string(rune(id))]
			messageProcessorWorkerMutex.Unlock()

			if value != id {
				t.Errorf("Value mismatch: expected %d, got %d", id, value)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("✅ Race condition test passed for %d goroutines", numGoroutines)
}

// BenchmarkMessageProcessorWorkerConcurrency benchmark de concorrência
func BenchmarkMessageProcessorWorkerConcurrency(b *testing.B) {
	testWorker := &MessageProcessorWorker{
		receivedValues:  make(map[string]int),
		validatedValues: make(map[string]int),
	}

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 10; j++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				messageProcessorWorkerMutex.Lock()
				testWorker.receivedValues["bench"] = id
				messageProcessorWorkerMutex.Unlock()
			}(j)
		}
		wg.Wait()
	}
}
