package certificate

type CertificatePack struct {
	CACertificateFilePath     string
	CAKeyFilePath             string
	ServerCertificateFilePath string
	ServerPrivateKeyFilePath  string
	ClientCertificateFilePath string
	ClientPrivateKeyFilePath  string
}

func NewCertificatePack() *CertificatePack {
	return &CertificatePack{}
}
