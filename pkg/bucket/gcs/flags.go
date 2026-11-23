package gcs

import (
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	namespace           = "gcs"
	credentialsFileFlag = namespace + ".credentials-file"
)

func RegisterFlags(namespace string, fs *pflag.FlagSet) {
	fs.String(flagKey(namespace, credentialsFileFlag), "", "Path to the credentials file")
}

// InitWithViper reads the flags that RegisterFlags registered under the same namespace.
func InitWithViper(namespace, bucket string, v *viper.Viper) (*Client, error) {
	credentialsFile := v.GetString(flagKey(namespace, credentialsFileFlag))
	return NewClient(bucket, WithCredentialsFile(credentialsFile))
}

func flagKey(namespace, key string) string {
	if namespace == "" {
		return key
	}
	return strings.TrimPrefix(namespace, ".") + "." + key
}
