package app

import (
	"fmt"
	"os"
	"strings"

	"kcl-lang.io/helm-kcl/pkg/helm"
)

const (
	DefaultHelmBinary             = "helm"
	DefaultKubeContext            = ""
	HelmRequiredVersion           = "v3.10.3"
	HelmRecommendedVersion        = "v3.11.2"
	HelmDiffRecommendedVersion    = "v3.4.0"
	HelmSecretsRecommendedVersion = "v4.1.1"
	HelmGitRecommendedVersion     = "v0.12.0"
	HelmS3RecommendedVersion      = "v0.14.0"
	HelmInstallCommand            = "https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3"
)

var (
	manuallyInstallCode   = 1
	windowPackageManagers = map[string]string{
		"scoop": fmt.Sprintf("scoop install helm@%s", strings.TrimLeft(HelmRecommendedVersion, "v")),
		"choco": fmt.Sprintf("choco install kubernetes-helm --version %s", strings.TrimLeft(HelmRecommendedVersion, "v")),
	}
	helmPlugins = []helmRecommendedPlugin{
		{
			name:    "diff",
			version: HelmDiffRecommendedVersion,
			repo:    "https://github.com/databus23/helm-diff",
		},
		{
			name:    "secrets",
			version: HelmSecretsRecommendedVersion,
			repo:    "https://github.com/jkroepke/helm-secrets",
		},
		{
			name:    "s3",
			version: HelmS3RecommendedVersion,
			repo:    "https://github.com/hypnoglow/helm-s3.git",
		},
		{
			name:    "helm-git",
			version: HelmGitRecommendedVersion,
			repo:    "https://github.com/aslafy-z/helm-git.git",
		},
	}
)

type helmRecommendedPlugin struct {
	name    string
	version string
	repo    string
}

func New() *App {
	return &App{helmBinary: DefaultHelmBinary, logger: NewLogger(os.Stdout, "debug"), render: helm.Render{}}
}
