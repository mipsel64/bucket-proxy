package bucket

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// writeServiceAccountKey writes a service-account key file that parses but cannot get a token.
func writeServiceAccountKey(t *testing.T, dir string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(map[string]string{
		"type":           "service_account",
		"project_id":     "test",
		"private_key_id": "0",
		"private_key":    string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		"client_email":   "test@test.iam.gserviceaccount.com",
		"client_id":      "1",
		"token_uri":      "https://oauth2.googleapis.com/token",
	})
	if err != nil {
		t.Fatal(err)
	}
	return writeFile(t, dir, "sa.json", string(b))
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// initFromArgs binds flags and env the same way as cmd/utils.go.
func initFromArgs(t *testing.T, args ...string) (Interface, error) {
	t.Helper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	RegisterFlags(fs)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	if err := v.BindPFlags(fs); err != nil {
		t.Fatal(err)
	}
	return InitFromViper(v)
}

func TestInitFromViper(t *testing.T) {
	dir := t.TempDir()
	validKey := writeServiceAccountKey(t, dir)
	malformedKey := writeFile(t, dir, "malformed.json", "{not json")
	missingKey := filepath.Join(dir, "missing.json")

	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		adc     string // GOOGLE_APPLICATION_CREDENTIALS
		wantErr string
	}{
		{
			name: "credentials file flag is used when ADC is unavailable",
			args: []string{"--bucket.name=b", "--bucket.gcs.credentials-file=" + validKey},
			adc:  missingKey,
		},
		{
			name:    "malformed credentials file flag fails even when ADC is valid",
			args:    []string{"--bucket.name=b", "--bucket.gcs.credentials-file=" + malformedKey},
			adc:     validKey,
			wantErr: "new google cloud storage client",
		},
		{
			name:    "missing credentials file flag fails even when ADC is valid",
			args:    []string{"--bucket.name=b", "--bucket.gcs.credentials-file=" + missingKey},
			adc:     validKey,
			wantErr: "new google cloud storage client",
		},
		{
			name: "BUCKET_GCS_CREDENTIALS_FILE env is used when ADC is unavailable",
			args: []string{"--bucket.name=b"},
			env:  map[string]string{"BUCKET_GCS_CREDENTIALS_FILE": validKey},
			adc:  missingKey,
		},
		{
			name: "ADC is used when no credentials file is set",
			args: []string{"--bucket.name=b"},
			adc:  validKey,
		},
		{
			name:    "no credentials file and no ADC fails",
			args:    []string{"--bucket.name=b"},
			adc:     missingKey,
			wantErr: "new google cloud storage client",
		},
		{
			name:    "missing bucket name fails",
			args:    []string{"--bucket.gcs.credentials-file=" + validKey},
			wantErr: "bucket name is required",
		},
		{
			name:    "unknown provider fails",
			args:    []string{"--bucket.name=b", "--bucket.provider=s3"},
			wantErr: `unknown provider "s3"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tt.adc)
			t.Setenv("BUCKET_GCS_CREDENTIALS_FILE", "")
			for k, val := range tt.env {
				t.Setenv(k, val)
			}

			b, err := initFromArgs(t, tt.args...)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("InitFromViper() error = %v, want nil", err)
				}
				if b == nil {
					t.Fatal("InitFromViper() returned nil bucket")
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("InitFromViper() error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
