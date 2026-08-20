package claimauditcommit

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestClaimAuditCommitOrdering(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestFailedClaimCommitDoesNotPublishAudit", false)
}
