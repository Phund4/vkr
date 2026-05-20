package services

import "testing"

func TestCrashAlertMetricValue(t *testing.T) {
	t.Parallel()
	const th = 0.5
	if got := crashAlertMetricValue(0.9, th); got != 1 {
		t.Fatalf("want 1 got %v", got)
	}
	if got := crashAlertMetricValue(0.5, th); got != 1 {
		t.Fatalf("want 1 at threshold got %v", got)
	}
	if got := crashAlertMetricValue(0.49, th); got != 0 {
		t.Fatalf("want 0 got %v", got)
	}
}
