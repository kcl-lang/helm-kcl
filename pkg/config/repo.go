package config

// RepositorySpec that defines values for a helm repo
type RepositorySpec struct {
	Name            string `yaml:"name,omitempty"`
	Path            string `yaml:"path,omitempty"`
	URL             string `yaml:"url,omitempty"`
	CaFile          string `yaml:"caFile,omitempty"`
	CertFile        string `yaml:"certFile,omitempty"`
	KeyFile         string `yaml:"keyFile,omitempty"`
	Username        string `yaml:"username,omitempty"`
	Password        string `yaml:"password,omitempty"`
	Managed         string `yaml:"managed,omitempty"`
	OCI             bool   `yaml:"oci,omitempty"`
	PassCredentials string `yaml:"passCredentials,omitempty"`
	SkipTLSVerify   string `yaml:"skipTLSVerify,omitempty"`
}
