package nconf

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

type TLSConfig struct {
	CAFiles  []string `mapstructure:"ca_files" envconfig:"ca_files"`
	KeyFile  string   `mapstructure:"key_file" split_words:"true"`
	CertFile string   `mapstructure:"cert_file" split_words:"true"`

	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
	CA   string `mapstructure:"ca"`

	Insecure bool `default:"false"`
	Enabled  bool `default:"false"`
}

func (cfg TLSConfig) TLSConfig() (*tls.Config, error) {
	var err error

	tlsConf := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: cfg.Insecure,
	}

	// Load CA
	if cfg.CA != "" {
		tlsConf.RootCAs, err = LoadCAFromValue(cfg.CA)
	} else if len(cfg.CAFiles) > 0 {
		tlsConf.RootCAs, err = LoadCAFromFiles(cfg.CAFiles)
	} else {
		tlsConf.RootCAs, err = x509.SystemCertPool()
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error setting up Root CA pool")
	}

	// Load Certs if any
	var cert tls.Certificate
	if cfg.Cert != "" && cfg.Key != "" {
		cert, err = LoadCertFromValues(cfg.Cert, cfg.Key)
		tlsConf.Certificates = append(tlsConf.Certificates, cert)
	} else if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err = LoadCertFromFiles(cfg.CertFile, cfg.KeyFile)
		tlsConf.Certificates = append(tlsConf.Certificates, cert)
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error loading certificate KeyPair")
	}

	// Backwards compatibility: if TLS is not explicitly enabled, return nil if no certificate was provided
	// Old code disabled TLS by not providing a certificate, which returned nil when calling TLSConfig()
	if !cfg.Enabled && len(tlsConf.Certificates) == 0 {
		return nil, nil
	}

	return tlsConf, nil
}

func LoadCertFromValues(certPEM, keyPEM string) (tls.Certificate, error) {
	return tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
}

func LoadCertFromFiles(certFile, keyFile string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(certFile, keyFile)
}

func LoadCAFromFiles(cafiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	for _, caFile := range cafiles {
		caData, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}

		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("Failed to add CA cert at %s", caFile)
		}
	}

	return pool, nil
}

func LoadCAFromValue(ca string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(ca)) {
		return nil, fmt.Errorf("Failed to add CA cert")
	}
	return pool, nil
}
