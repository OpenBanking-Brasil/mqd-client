package application

import (
	"sync"
	"testing"
	"time"

	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting"
	"github.com/OpenBanking-Brasil/MQD_Client/crosscutting/log"
)

// TestResultProcessorTotalResultsRace reproduz, de forma isolada, a data race que já
// existe hoje em result_processor.go entre:
//   - ResultProcessor.AppendResult, que incrementa a variavel de pacote `totalResults`
//     protegida por `resultProcessorMutex` (chamada pela goroutine do worker); e
//   - a checagem de limiar dentro do select de StartResultsProcessor
//     ("if totalResults >= rp.cm.GetSendOnReportNumber()"), que le a MESMA variavel
//     SEM segurar esse mutex (roda em outra goroutine).
//
// Nenhum codigo de producao foi alterado. Este arquivo apenas evidencia o problema
// usando exatamente o mesmo estado compartilhado do pacote.
//
// Rodar com o race detector:
//   go test -race -run TestResultProcessorTotalResultsRace ./application/... -v
func TestResultProcessorTotalResultsRace(t *testing.T) {
	logger := log.GetNewJSONLogger()
	logger.SetLoggingGlobalLevel(log.Disabled)

	rp := &ResultProcessor{
		OFBStruct: crosscutting.OFBStruct{Pack: "ResultProcessor", Logger: logger},
		cm:        &ConfigurationManager{},
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: mesma escrita real de producao (AppendResult, com lock)
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			rp.AppendResult(&MessageResult{ServerID: "server-1", TransmitterID: "tx-1"})
		}
	}()

	// Goroutine 2: mesma leitura desprotegida que existe em StartResultsProcessor
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			_ = totalResults >= 10 // leitura sem mutex, identica ao codigo de producao
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
}
