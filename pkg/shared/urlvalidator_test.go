package shared

import "testing"

func TestAccepedtPort(t *testing.T) {
	t.Parallel()
	for _, p := range []int{80, 443} {
		if !AcceptedPort(p) {
			t.Errorf("port %d should be allowed", p)
		}
	}
	for _, p := range []int{1, 6000, 179} {
		if AcceptedPort(p) {
			t.Errorf("port %d should not be allowed", p)
		}
	}
}
