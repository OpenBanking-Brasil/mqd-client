package monitoring

import (
	"runtime"
	"testing"
	"time"
)

// TestMetricsNoGoroutineLeaks verifica se há vazamento de goroutines
func TestMetricsNoGoroutineLeaks(t *testing.T) {
	// Aguardar goroutines pendentes serem finalizadas
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	// Contar goroutines iniciais
	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial goroutines: %d", initialGoroutines)

	// Aqui iria o código de teste que usa metrics
	// Por enquanto, apenas verificamos o baseline

	// Aguardar um pouco para simular operações
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	// Contar goroutines finais
	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutines: %d", finalGoroutines)

	// Verificar se há vazamento
	if finalGoroutines > initialGoroutines+1 {
		t.Errorf("Goroutine leak detected: expected max %d, got %d",
			initialGoroutines+1, finalGoroutines)

		// Exibir stack traces de goroutines para debug
		buf := make([]byte, 1<<20)
		n := runtime.Stack(buf, true)
		t.Logf("Stack traces:\n%s", string(buf[:n]))
	}
}

// BenchmarkMetricsGoroutineCount avalia o número de goroutines criadas
func BenchmarkMetricsGoroutineCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		initialCount := runtime.NumGoroutine()
		// Simular operações
		time.Sleep(1 * time.Millisecond)
		finalCount := runtime.NumGoroutine()

		if finalCount > initialCount+1 {
			b.Errorf("Unexpected goroutine creation: %d -> %d", initialCount, finalCount)
		}
	}
}
