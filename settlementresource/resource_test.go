package settlementresource

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestSettlementResourceLifetime(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestLargeSettlementClosesEachResourcePromptly", false)
}
