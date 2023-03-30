package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"k8s.io/helm/pkg/helm"
)

var (
	helmVersionRE  = regexp.MustCompile(`Version:\s*"([^"]+)"`)
	minHelmVersion = semver.MustParse("v3.1.0-rc.1")
)

// HelmBin returns the helm binary path.
func HelmBin() string {
	helmBin := DefaultHelmBinary
	if os.Getenv("HELM_BIN") != "" {
		helmBin = os.Getenv("HELM_BIN")
	}
	return helmBin
}

func compatibleHelm3Version() error {
	cmd := exec.Command(HelmBin(), "version")
	debugPrint("Executing %s", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to run `%s version`: %v", HelmBin(), err)
	}
	versionOutput := string(output)

	matches := helmVersionRE.FindStringSubmatch(versionOutput)
	if matches == nil {
		return fmt.Errorf("Failed to find version in output %#v", versionOutput)
	}
	helmVersion, err := semver.NewVersion(matches[1])
	if err != nil {
		return fmt.Errorf("Failed to parse version %#v: %v", matches[1], err)
	}

	if minHelmVersion.GreaterThan(helmVersion) {
		return fmt.Errorf("helm diff upgrade requires at least helm version %s", minHelmVersion.String())
	}
	return nil

}
func getRelease(release, namespace string) ([]byte, error) {
	args := []string{"get", "manifest", release}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	cmd := exec.Command(HelmBin(), args...)
	return outputWithRichError(cmd)
}

func getHooks(release, namespace string) ([]byte, error) {
	args := []string{"get", "hooks", release}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	cmd := exec.Command(HelmBin(), args...)
	return outputWithRichError(cmd)
}

func getRevision(release string, revision int, namespace string) ([]byte, error) {
	args := []string{"get", "manifest", release, "--revision", strconv.Itoa(revision)}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	cmd := exec.Command(HelmBin(), args...)
	return outputWithRichError(cmd)
}

func getChart(release, namespace string) (string, error) {
	args := []string{"get", "all", release, "--template", "{{.Release.Chart.Name}}"}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	cmd := exec.Command(HelmBin(), args...)
	out, err := outputWithRichError(cmd)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

type executor struct {
	release                  string
	chart                    string
	chartVersion             string
	chartRepo                string
	client                   helm.Interface
	detailedExitCode         bool
	devel                    bool
	disableValidation        bool
	disableOpenAPIValidation bool
	dryRun                   bool
	namespace                string // namespace to assume the release to be installed into. Defaults to the current kube config namespace.
	valueFiles               valueFiles
	values                   []string
	stringValues             []string
	fileValues               []string
	reuseValues              bool
	resetValues              bool
	allowUnreleased          bool
	noHooks                  bool
	includeTests             bool
	postRenderer             string
	install                  bool
	normalizeManifests       bool
	extraAPIs                []string
	kubeVersion              string
	useUpgradeDryRun         bool
	isAllowUnreleased        bool
}

func (e *executor) template(isUpgrade bool) ([]byte, error) {
	flags := []string{}
	if e.devel {
		flags = append(flags, "--devel")
	}
	if e.noHooks && !e.useUpgradeDryRun {
		flags = append(flags, "--no-hooks")
	}
	if e.chartVersion != "" {
		flags = append(flags, "--version", e.chartVersion)
	}
	if e.chartRepo != "" {
		flags = append(flags, "--repo", e.chartRepo)
	}
	if e.namespace != "" {
		flags = append(flags, "--namespace", e.namespace)
	}
	if e.postRenderer != "" {
		flags = append(flags, "--post-renderer", e.postRenderer)
	}

	shouldDefaultReusingValues := isUpgrade && len(e.values) == 0 && len(e.stringValues) == 0 && len(e.valueFiles) == 0 && len(e.fileValues) == 0
	if (e.reuseValues || shouldDefaultReusingValues) && !e.resetValues && !e.dryRun {
		tmpfile, err := ioutil.TempFile("", "existing-values")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpfile.Name())
		if err := e.writeExistingValues(tmpfile); err != nil {
			return nil, err
		}
		flags = append(flags, "--values", tmpfile.Name())
	}
	for _, value := range e.values {
		flags = append(flags, "--set", value)
	}
	for _, stringValue := range e.stringValues {
		flags = append(flags, "--set-string", stringValue)
	}
	for _, valueFile := range e.valueFiles {
		if strings.TrimSpace(valueFile) == "-" {
			bytes, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return nil, err
			}

			tmpfile, err := ioutil.TempFile("", "helm-kcl-stdin-values")
			if err != nil {
				return nil, err
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write(bytes); err != nil {
				tmpfile.Close()
				return nil, err
			}

			if err := tmpfile.Close(); err != nil {
				return nil, err
			}

			flags = append(flags, "--values", tmpfile.Name())
		} else {
			flags = append(flags, "--values", valueFile)
		}
	}
	for _, fileValue := range e.fileValues {
		flags = append(flags, "--set-file", fileValue)
	}

	if e.disableOpenAPIValidation {
		flags = append(flags, "--disable-openapi-validation")
	}

	var (
		subcmd string
		filter func([]byte) []byte
	)

	if e.useUpgradeDryRun {
		if e.dryRun {
			return nil, fmt.Errorf("`diff upgrade --dry-run` conflicts with HELM_DIFF_USE_UPGRADE_DRY_RUN_AS_TEMPLATE. Either remove --dry-run to enable cluster access, or unset HELM_DIFF_USE_UPGRADE_DRY_RUN_AS_TEMPLATE to make cluster access unnecessary")
		}

		if e.isAllowUnreleased {
			// Otherwise you get the following error when this is a diff for a new install
			//   Error: UPGRADE FAILED: "$RELEASE_NAME" has no deployed releases
			flags = append(flags, "--install")
		}

		flags = append(flags, "--dry-run")
		subcmd = "upgrade"
		filter = func(s []byte) []byte {
			return extractManifestFromHelmUpgradeDryRunOutput(s, e.noHooks)
		}
	} else {
		if !e.disableValidation && !e.dryRun {
			flags = append(flags, "--validate")
		}

		if isUpgrade {
			flags = append(flags, "--is-upgrade")
		}

		for _, a := range e.extraAPIs {
			flags = append(flags, "--api-versions", a)
		}

		if e.kubeVersion != "" {
			flags = append(flags, "--kube-version", e.kubeVersion)
		}

		subcmd = "template"

		filter = func(s []byte) []byte {
			return s
		}
	}

	args := []string{subcmd, e.release, e.chart}
	args = append(args, flags...)

	cmd := exec.Command(HelmBin(), args...)
	out, err := outputWithRichError(cmd)
	return filter(out), err
}

func (e *executor) writeExistingValues(f *os.File) error {
	cmd := exec.Command(HelmBin(), "get", "values", e.release, "--all", "--output", "yaml")
	debugPrint("Executing %s", strings.Join(cmd.Args, " "))
	defer f.Close()
	cmd.Stdout = f
	return cmd.Run()
}

func extractManifestFromHelmUpgradeDryRunOutput(s []byte, noHooks bool) []byte {
	if len(s) == 0 {
		return s
	}

	i := bytes.Index(s, []byte("HOOKS:"))
	hooks := s[i:]

	j := bytes.Index(hooks, []byte("MANIFEST:"))

	manifest := hooks[j:]
	hooks = hooks[:j]

	k := bytes.Index(manifest, []byte("\nNOTES:"))

	if k > -1 {
		manifest = manifest[:k+1]
	}

	if noHooks {
		hooks = nil
	} else {
		a := bytes.Index(hooks, []byte("---"))
		if a > -1 {
			hooks = hooks[a:]
		} else {
			hooks = nil
		}
	}

	a := bytes.Index(manifest, []byte("---"))
	if a > -1 {
		manifest = manifest[a:]
	}

	r := []byte{}
	r = append(r, manifest...)
	r = append(r, hooks...)

	return r
}
