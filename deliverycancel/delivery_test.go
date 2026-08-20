package deliverycancel

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestDeliveryCancellationBoundary(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestCancelledDeliveryStopsRetriesAndNextRequestIsClean", false)
}
