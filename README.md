# Helm KCL Plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/helm-kcl)](https://goreportcard.com/report/github.com/KusionStack/helm-kcl)
[![GoDoc](https://godoc.org/github.com/KusionStack/helm-kcl?status.svg)](https://godoc.org/github.com/KusionStack/helm-kcl)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/KusionStack/helm-kcl/blob/main/LICENSE)

[KCL](https://github.com/KusionStack/KCLVM) is a constraint-based record & functional domain language. Full documents of KCL can be found [here](https://kcl-lang.io/).

You can use the `Helm-KCL-Plugin` to

+ Modify the `values.yaml` value of the helm charts you rely on by certain conditions or dynamically.
+ Edit the helm charts in a hook way to separate data and logic for the Kubernetes manifests management.
+ For multi-environment and multi-tenant scenarios, you can maintain these configurations gracefully rather than simply copy and paste.
+ Validate all KRM resources using the KCL schema.

## Prerequisites

+ Install Helm
+ Golang (at least version 1.18)

## Test the Plugin

You need to put your KCL script source in the functionConfig of kind KCLRun and then the function will run the KCL script that you provide.

```bash
# Verify that the annotation is added to the `Deployment` resource and the other resource `Service` 
# does not have this annotation.
diff \
  <(helm template ./examples/workload-charts-with-kcl/workload-charts) \
  <(go run main.go template --file ./examples/workload-charts-with-kcl/kcl-run.yaml) |\
  grep annotations -A1
```

The output is

```diff
>   annotations:
>     managed-by: helm-kcl-plugin
```

## Install

### Using Helm plugin manager (> 2.3.x)

```shell
helm plugin install https://github.com/KusionStack/helm-kcl
```

### Pre Helm 2.3.0 Installation

Pick a release tarball from the [releases page](https://github.com/KusionStack/helm-kcl/releases).

Unpack the tarball in your helm plugins directory ($(helm home)/plugins).

E.g.

```shell
curl -L $TARBALL_URL | tar -C $(helm home)/plugins -xzv
```

## From Source

### Prerequisites

+ GoLang 1.18+

Make sure you do not have a version of `helm-kcl` installed. You can remove it by running the command.

```shell
helm plugin uninstall kcl
```

### Installation Steps

The first step is to download the repository and enter the directory. You can do this via git clone or downloading and extracting the release. If you clone via git, remember to check out the latest tag for the latest release.

Next, depending on which helm version you have, install the plugin into helm.

#### Helm 2

```shell
make install
```

#### Helm 3

```shell
make install/helm3
```

## Build

### Prerequisites

+ GoLang 1.18+

```shell
git clone https://github.com/KusionStack/helm-kcl.git
cd helm-kcl
go run main.go
```

## Test

```shell
go test -v ./...
```

## Release

Bump version in `plugin.yaml`:

```shell
code plugin.yaml
git commit -m 'Bump helm-diff version to 0.x.y'
```

Set `GITHUB_TOKEN` and run:

```shell
make docker-run-release
```
