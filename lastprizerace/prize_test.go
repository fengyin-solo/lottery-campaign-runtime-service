package lastprizerace

import (
	"testing"

	"lottery/internal/testbridge"
)

func TestLastPrizeReservationRace(t *testing.T) {
	testbridge.RunRuntimeFlowTest(t, "TestConcurrentLastPrizeProducesOneWinner", true)
}
