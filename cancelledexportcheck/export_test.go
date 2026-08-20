package cancelledexportcheck

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestCancelledExportLifecycle(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestCancelledExportStopsAndNextExportIsolated", false)
}
