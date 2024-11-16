package credentials

import "errors"

var InvalidHostname = errors.New("server hostname cannot be blank")
var InvalidCertificateBasePath = errors.New("certificate base path invalid")

type CertificateGenerator interface {
	Generate() (*CertificatePack, error)
}
