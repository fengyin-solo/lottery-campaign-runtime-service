package testbridge

import (
	"os/exec"
	"testing"
)

func RunRuntimeFlowTest(t *testing.T, name string, race bool) {
	t.Helper()
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	args = append(args, "lottery/internal/runtimeflow/service", "-run", "^"+name+"$", "-count=1")
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("runtime flow scenario failed: %v\n%s", err, output)
	}
}
