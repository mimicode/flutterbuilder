package types

// IOSConfig iOS构建配置
type IOSConfig struct {
	P12Cert             string // P12证书文件路径
	CertPassword        string // 证书密码
	ProvisioningProfile string // 描述文件路径
	TeamID              string // 开发者团队ID
	BundleID            string // 应用Bundle ID
}

// CertificateManager iOS证书管理器接口
type CertificateManager interface {
	SetupCertificates() error
	CleanupCertificates() error
	CreateExportOptionsPlist() (string, error)
}
