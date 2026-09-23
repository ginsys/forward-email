package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/spf13/viper"

	"github.com/ginsys/forward-email/internal/testutil"
)

// TestAuthVerifyProfileSelection runs the real `auth verify` command against a fake API
// and checks which profile's API key reached the server, so a pass cannot come from a
// flag-parse error or a missing credential alone (issue #50: `-p` was rejected).
func TestAuthVerifyProfileSelection(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantKey string
	}{
		{name: "no flag uses current profile", args: []string{"auth", "verify"}, wantKey: "key-main"},
		{name: "long flag selects profile", args: []string{"auth", "verify", "--profile", "other"}, wantKey: "key-other"},
		{name: "shorthand flag selects profile", args: []string{"auth", "verify", "-p", "other"}, wantKey: "key-other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mu sync.Mutex
			var gotKeys []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				key, _, _ := r.BasicAuth()
				mu.Lock()
				gotKeys = append(gotKeys, key)
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("[]"))
			}))
			defer srv.Close()

			isolateAuthVerify(t, srv.URL)

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&out)
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()

			mu.Lock()
			defer mu.Unlock()
			if len(gotKeys) == 0 {
				t.Fatalf("no request reached the API (err=%v, output=%q)", err, out.String())
			}
			for _, k := range gotKeys {
				if k != tt.wantKey {
					t.Errorf("API received key %q, want %q (profile selection)", k, tt.wantKey)
				}
			}
			if err != nil {
				t.Errorf("auth verify returned error: %v (output=%q)", err, out.String())
			}
		})
	}
}

// isolateAuthVerify points config at a disposable directory with two profiles, disables the
// OS keyring, clears credential environment variables and resets command/viper state so the
// test never reads the developer's real configuration or keyring.
func isolateAuthVerify(t *testing.T, baseURL string) {
	t.Helper()

	tempDir := testutil.SetupTempConfigWithReset(t)
	t.Setenv("HOME", tempDir)
	testutil.WriteTestConfig(t, tempDir, `current_profile: "main"
profiles:
  main:
    api_key: "key-main"
  other:
    api_key: "key-other"
`)
	t.Setenv("FORWARDEMAIL_KEYRING_BACKEND", "none")
	for _, name := range []string{"FORWARDEMAIL_API_KEY", "FORWARDEMAIL_MAIN_API_KEY",
		"FORWARDEMAIL_OTHER_API_KEY", "FORWARDEMAIL_PROFILE", "FORWARDEMAIL_CURRENT_PROFILE",
		"FORWARDEMAIL_API_BASE_URL"} {
		t.Setenv(name, "")
	}
	viper.Set("api_base_url", baseURL)

	resetProfileFlags := func() {
		if f := authVerifyCmd.Flags().Lookup("profile"); f != nil {
			_ = f.Value.Set("")
			f.Changed = false
		}
		if f := rootCmd.PersistentFlags().Lookup("profile"); f != nil {
			_ = f.Value.Set("")
			f.Changed = false
		}
	}
	resetProfileFlags()
	t.Cleanup(func() {
		resetProfileFlags()
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		viper.Reset()
	})
}
