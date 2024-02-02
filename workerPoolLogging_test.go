package workerpool

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWorkerPoolZapLogger(t *testing.T) {
	maxWorkers := 4
	taskQueueSize := 100
	timeout := 10 * time.Second
	taskNums := 100

	zapLogger, _ := NewZapLogger()
	workerPoolConfig := WorkerPoolConfig{
		MaxWorkers:    maxWorkers,
		Timeout:       timeout,
		TaskQueueSize: taskQueueSize,
		Logger:        zapLogger,
	}
	workerPool := NewWorkerPool(workerPoolConfig)
	task := func() (interface{}, error) {
		return "zap logger", nil
	}

	time.Sleep(1 * time.Second)
	for i := 0; i < taskNums; i++ {
		workerPool.Submit(task)
	}

	// this is blocking
	for i := 0; i < taskNums; i++ {
		taskResult := <-workerPool.resultChan
		if taskResult.Error != nil {
			t.Error(taskResult.Error)
		}
	}
	workerPool.Stop()
}

// ----------------------------------------------------------------
// Zap Helpers
type ZapLogger struct {
	zapLogger *zap.Logger
}

func NewZapLogger() (*ZapLogger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()
	return &ZapLogger{zapLogger: logger}, nil
}

func (l *ZapLogger) Log(message string) {
	l.zapLogger.Info(message)
}
