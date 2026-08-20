package wrappedretrycheck

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestWrappedRetryClassification(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestWrappedTemporaryRedemptionRetriesWithoutDuplicateCommit", false)
}
