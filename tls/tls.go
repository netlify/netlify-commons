package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

type Config struct {
	CAFiles  []string `mapstructure:"ca_files" envconfig:"ca_files"`
	KeyFile  string   `mapstructure:"key_file" split_words:"true"`
	CertFile string   `mapstructure:"cert_file" split_words:"true"`

	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
	CA   string `mapstructure:"ca"`

	Insecure bool `default:"false"`
}

func (cfg Config) TLSConfig() (*tls.Config, error) {
	var tlsconf *tls.Config
	var err error
	if cfg.Cert != "" && cfg.Key != "" {
		tlsconf, err = LoadFromValues(cfg.Cert, cfg.Key, cfg.CA)
	} else if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsconf, err = LoadFromFiles(cfg.CertFile, cfg.KeyFile, cfg.CAFiles)
	}

	if err != nil {
		return nil, err
	}

	if tlsconf != nil {
		tlsconf.InsecureSkipVerify = cfg.Insecure
	}

	return tlsconf, nil
}

func LoadFromValues(certPEM, keyPEM, ca string) (*tls.Config, error) {
	var pool *x509.CertPool
	// If no CA cert if provided, use system pool
	if ca == "" {
		p, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		pool = p
	} else {
		pool = x509.NewCertPool()
		if !pool.AppendCertsFromPEM([]byte(ca)) {
			return nil, fmt.Errorf("Failed to add CA cert")
		}
	}

	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

func LoadFromFiles(certFile, keyFile string, cafiles []string) (*tls.Config, error) {
	var pool *x509.CertPool
	if len(cafiles) == 0 {
		p, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		pool = p
	} else {
		pool = x509.NewCertPool()

		for _, caFile := range cafiles {
			caData, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, err
			}

			if !pool.AppendCertsFromPEM(caData) {
				return nil, fmt.Errorf("Failed to add CA cert at %s", caFile)
			}
		}
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}
