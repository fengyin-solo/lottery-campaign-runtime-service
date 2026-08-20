package pooledidentity

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestPooledRequestIdentityIsolation(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestPooledAuditDoesNotLeakPreviousIdentity", false)
}
