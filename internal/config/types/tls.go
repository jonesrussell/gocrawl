package types

import "errors"

// TLSConfig holds TLS configuration settings.
type TLSConfig struct {
	// Enabled indicates whether TLS is enabled
	Enabled bool `yaml:"enabled"`
	// CertFile is the path to the certificate file
	CertFile string `yaml:"cert_file"`
	// KeyFile is the path to the key file
	KeyFile string `yaml:"key_file"`
	// CAFile is the path to the CA certificate file
	CAFile string `yaml:"ca_file"`
	// InsecureSkipVerify indicates whether to skip certificate verification
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// Validate validates the TLS configuration.
func (c *TLSConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if !c.InsecureSkipVerify {
		if c.CertFile == "" {
			return errors.New("cert_file is required when TLS is enabled and insecure_skip_verify is false")
		}
		if c.KeyFile == "" {
			return errors.New("key_file is required when TLS is enabled and insecure_skip_verify is false")
		}
	}
	return nil
}
