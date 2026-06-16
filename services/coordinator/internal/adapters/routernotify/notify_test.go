package routernotify

import "testing"

func TestReloadURL(t *testing.T) {
	t.Parallel()
	if got := ReloadURL("http://router:9091/metrics"); got != "http://router:9091/v1/reload" {
		t.Fatalf("got %q", got)
	}
	if got := ReloadURL("http://router:9091/"); got != "http://router:9091/v1/reload" {
		t.Fatalf("got %q", got)
	}
}
