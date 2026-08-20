package typednilnotice

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestTypedNilNotificationLifecycle(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestTypedNilNotifierSkipsAndLaterSendStillWorks", false)
}
