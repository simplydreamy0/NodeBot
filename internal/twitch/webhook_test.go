package twitch

import (
	"testing"
)

func TestVerifyHMAC(t *testing.T) {
	secret := "veryverysecuresecret"
	msg := "this is a test message"

	tests := []struct {
		name			string
		sig       string
		wants			bool
	}{
		{
			name: "true signature",
			sig: "d314ca20aee4965a99394003219dec478327bc669e9c52da84fd3d80730d80ad",
			wants: true,
		},
		{
			name: "false signature",
			sig: "d314ca20aee4965a99394003219dec478327bc669e9c52da84fd3d8ogfzrofz",
			wants: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := verifyHMAC(msg, secret, tt.sig); got != tt.wants {
				t.Errorf("Signature: %s wants: %v, but got %v", tt.sig, tt.wants, got)
			}
		})
	}
}
