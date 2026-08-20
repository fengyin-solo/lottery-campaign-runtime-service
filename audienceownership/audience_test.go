package audienceownership

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestSubmittedAudienceOwnership(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestAudienceBatchKeepsSubmittedUsers", false)
}
