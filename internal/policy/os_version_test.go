package policy

import (
	"testing"
)

func TestOSVersion_Policy(t *testing.T) {
	// Save original function
	origFunc := getCurrentWindowsVersion
	defer func() { getCurrentWindowsVersion = origFunc }()

	tests := []struct {
		name       string
		mockOS     string
		policyVers []string
		shouldPass bool
	}{
		{
			name:       "Exact Match Server 2025",
			mockOS:     "Server 2025",
			policyVers: []string{"Server 2025"},
			shouldPass: true,
		},
		{
			name:       "Mismatch Server 2019 on 2025",
			mockOS:     "Server 2025",
			policyVers: []string{"Server 2019"},
			shouldPass: false,
		},
		{
			name:       "Mismatch Server 2025 on 2019",
			mockOS:     "Server 2019",
			policyVers: []string{"Server 2025"},
			shouldPass: false,
		},
		{
			name:       "Multiple Allowed",
			mockOS:     "Server 2022",
			policyVers: []string{"Server 2019", "Server 2022"},
			shouldPass: true,
		},
		{
			name:       "Case Insensitive",
			mockOS:     "server 2025",
			policyVers: []string{"SERVER 2025"},
			shouldPass: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mock the OS version
			getCurrentWindowsVersion = func() string {
				return tc.mockOS
			}

			p := &OSVersionPolicy{
				OS:              "windows", // Assuming test runs where runtime.GOOS matches or we mock runtime check?
				AllowedVersions: tc.policyVers,
			}
			// Note: If running on linux, p.Check() checks runtime.GOOS == "windows".
			// We might need to bypass that check or set p.OS = runtime.GOOS for the test.
			// But since we are testing verification logic, we can just set p.OS = "" to skip OS type check
			// or assume we are on windows (which CI is).
			// To be safe on Linux dev env:
			p.OS = ""

			err := p.Check()
			if tc.shouldPass && err != nil {
				t.Errorf("Expected pass, got error: %v", err)
			}
			if !tc.shouldPass && err == nil {
				t.Errorf("Expected failure, got nil")
			}
		})
	}
}
