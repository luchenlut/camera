package camera

import (
  "crypto/tls"
  "io/ioutil"
  "crypto/x509"
)

// NewTLSConfig  the TLS configuration.
func NewTLSConfig(caFile, certFile, certKeyFile string) (*tls.Config, error) {
  if caFile == "" && certFile == "" && certKeyFile == "" {
    return nil, nil
  }

  tlsConfig := &tls.Config{}

  if caFile != "" {
    caCert, err := ioutil.ReadFile(caFile)
    if err != nil {
      return nil, err
    }
    certPool := x509.NewCertPool()
    certPool.AppendCertsFromPEM(caCert)
    tlsConfig.RootCAs = certPool
  }
  tlsConfig.InsecureSkipVerify = true
  if certFile != "" && certKeyFile != "" {
    kp, err := tls.LoadX509KeyPair(certFile, certKeyFile)
    if err != nil {
      return nil, err
    }
    tlsConfig.Certificates = []tls.Certificate{kp}
  }

  return tlsConfig, nil
}