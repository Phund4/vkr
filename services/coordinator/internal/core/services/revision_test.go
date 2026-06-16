package services

import "testing"

func TestBumpAssignmentsRevision(t *testing.T) {
	t.Parallel()
	before := CurrentAssignmentsRevision()
	after := BumpAssignmentsRevision()
	if after != before+1 {
		t.Fatalf("want %d got %d", before+1, after)
	}
	if CurrentAssignmentsRevision() != after {
		t.Fatalf("current revision mismatch")
	}
}
