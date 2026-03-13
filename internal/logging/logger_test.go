package logging

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		appEnv  string
		wantErr bool
	}{
		{name: "development", appEnv: "development"},
		{name: "production", appEnv: "production"},
		{name: "invalid", appEnv: "staging", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger, err := New(tc.appEnv)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for env %q", tc.appEnv)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if logger == nil {
				t.Fatalf("expected logger for env %q", tc.appEnv)
			}
		})
	}
}
