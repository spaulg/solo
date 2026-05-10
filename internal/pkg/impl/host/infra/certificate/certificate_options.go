package certificate

import (
	"crypto/x509"
	"time"
)

type Option func(template *x509.Certificate)

func WithDuration(duration time.Duration) Option {
	return func(template *x509.Certificate) {
		template.NotAfter = time.Now().Add(duration)
	}
}

func WithKeyUsage(keyUsage x509.KeyUsage) Option {
	return func(template *x509.Certificate) {
		template.KeyUsage = keyUsage
	}
}

func WithExtKeyUsage(extKeyUsage []x509.ExtKeyUsage) Option {
	return func(template *x509.Certificate) {
		template.ExtKeyUsage = extKeyUsage
	}
}

func WithOrganization(organization []string) Option {
	return func(template *x509.Certificate) {
		template.Subject.Organization = organization
	}
}

func WithCommonName(commonName string) Option {
	return func(template *x509.Certificate) {
		template.Subject.CommonName = commonName
	}
}

func WithDNSNames(dnsNames []string) Option {
	return func(template *x509.Certificate) {
		template.DNSNames = dnsNames
	}
}
