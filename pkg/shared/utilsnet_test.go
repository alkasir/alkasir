// +build net

package shared

import (
	"bytes"
	"testing"
)

func TestResolveWAN(t *testing.T) {
	ip := GetPublicIPAddr()
	ip2 := GetPublicIPAddr()
	if !bytes.Equal(ip, ip2) {
		t.Fatalf("%s and %s are not equel", ip.String(), ip2.String())
	}
}
