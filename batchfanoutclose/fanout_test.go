package batchfanoutclose

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestBatchFanoutCompletion(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestBatchDrawReturnsEveryTaskBeforeClosing", true)
}
