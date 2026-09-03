package security

import (
	"testing"
)

func TestValidateTargetURL(t *testing.T) {
	tests := []struct {
		url          string
		allowPrivate bool
		wantErr      bool
	}{
		// Valid public URLs
		{"https://example.com", false, false},
		{"http://news.ycombinator.com/item?id=1", false, false},
		{"https://github.com", false, false},

		// Schemes
		{"file:///etc/passwd", false, true},
		{"ftp://example.com", false, true},
		{"gopher://example.com", false, true},
		{"javascript:alert(1)", false, true},

		// Restricted loopback / localhost
		{"http://127.0.0.1:8080", false, true},
		{"http://localhost:3000", false, true},
		{"http://sub.localhost:8080", false, true},

		// Cloud metadata & link-local
		{"http://169.254.169.254/latest/meta-data/", false, true},
		{"http://169.254.1.1", false, true},

		// Private RFC 1918 networks
		{"http://10.0.0.1/admin", false, true},
		{"http://192.168.1.1/setup", false, true},
		{"http://172.16.0.1", false, true},

		// Allowed when allowPrivate = true
		{"http://127.0.0.1:8080", true, false},
		{"http://localhost:3000", true, false},
		{"http://192.168.1.1", true, false},
	}

	for _, tt := range tests {
		err := ValidateTargetURL(tt.url, tt.allowPrivate)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateTargetURL(%q, allowPrivate=%v) err = %v, wantErr = %v", tt.url, tt.allowPrivate, err, tt.wantErr)
		}
	}
}
