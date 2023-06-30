# Helm KCL Plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/kcl-lang/helm-kcl)](https://goreportcard.com/report/github.com/kcl-lang/helm-kcl)
[![GoDoc](https://godoc.org/github.com/kcl-lang/helm-kcl?status.svg)](https://godoc.org/github.com/kcl-lang/helm-kcl)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kcl-lang/helm-kcl/blob/main/LICENSE)

[KCL](https://github.com/KusionStack/kcl) is a constraint-based record & functional domain language. Full documents of KCL can be found [here](https://kcl-lang.io/).

You can use the `Helm-KCL-Plugin` to

+ Edit the helm charts in a hook way to separate data and logic for the Kubernetes manifests management.
+ For multi-environment and multi-tenant scenarios, you can maintain these configurations gracefully rather than simply copy and paste.
+ Validate all KRM resources using the KCL schema.

## Install

### Using Helm plugin manager (> 2.3.x)

```shell
helm plugin install https://github.com/kcl-lang/helm-kcl
```

### Pre Helm 2.3.0 Installation

Pick a release tarball from the [releases page](https://github.com/kcl-lang/helm-kcl/releases).

Unpack the tarball in your helm plugins directory ($(helm home)/plugins).

E.g.

```shell
curl -L $TARBALL_URL | tar -C $(helm home)/plugins -xzv
```

### Install From Source

#### Prerequisites

+ GoLang 1.18+

Make sure you do not have a version of `helm-kcl` installed. You can remove it by running the command.

```shell
helm plugin uninstall kcl
```

#### Installation Steps

The first step is to download the repository and enter the directory. You can do this via git clone or downloading and extracting the release. If you clone via git, remember to check out the latest tag for the latest release.

Next, depending on which helm version you have, install the plugin into helm.

##### Helm 2

```shell
make install
```

##### Helm 3

```shell
make install/helm3
```

## Quick Start

```shell
helm kcl template --file ./examples/workload-charts-with-kcl/kcl-run.yaml
```

The content of `kcl-run.yaml` looks like this:

```yaml
# kcl-config.yaml
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: set-annotation
spec:
  # EDIT THE SOURCE!
  # This should be your KCL code which preloads the `ResourceList` to `option("resource_list")
  source: |
    [resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "helm-kcl-plugin"}} for resource in option("resource_list").items]

repositories:
  - name: workload
    path: ./workload-charts
```

The output is:

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/instance: workload
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: workload
    app.kubernetes.io/version: 0.1.0
    helm.sh/chart: workload-0.1.0
  name: workload
spec:
  ports:
  - name: www
    port: 80
    protocol: TCP
    targetPort: 80
  selector:
    app.kubernetes.io/instance: workload
    app.kubernetes.io/name: workload
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/instance: workload
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: workload
    app.kubernetes.io/version: 0.1.0
    helm.sh/chart: workload-0.1.0
  name: workload
  annotations:
    managed-by: helm-kcl-plugin
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: workload
      app.kubernetes.io/name: workload
  template:
    metadata:
      labels:
        app.kubernetes.io/instance: workload
        app.kubernetes.io/name: workload
    spec:
      containers:
      - image: "nginx:alpine"
        name: frontend
```

## Build

### Prerequisites

+ GoLang 1.18+

```shell
git clone https://github.com/kcl-lang/helm-kcl.git
cd helm-kcl
go run main.go
```

## Test

### Unit Test

```shell
go test -v ./...
```

### Integration Test

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

## Release

Bump version in `plugin.yaml`:

```shell
code plugin.yaml
git commit -m 'Bump helm-kcl version to 0.x.y'
```

Set `GITHUB_TOKEN` and run:

```shell
make docker-run-release
```

## Guides for Developing KCL

Here's what you can do in the KCL script:

+ Read resources from `option("resource_list")`. The `option("resource_list")` complies with the [KRM Functions Specification](https://kpt.dev/book/05-developing-functions/01-functions-specification). You can read the input resources from `option("resource_list")["items"]` and the `functionConfig` from `option("resource_list")["functionConfig"]`.
+ Return a KPM list for output resources.
+ Return an error using `assert {condition}, {error_message}`.
+ Read the environment variables. e.g. `option("PATH")` (Not yet implemented).
+ Read the OpenAPI schema. e.g. `option("open_api")["definitions"]["io.k8s.api.apps.v1.Deployment"]` (Not yet implemented).

Full documents of KCL can be found [here](https://kcl-lang.io/).

## Examples

See [here](https://kcl-lang.io/krm-kcl/tree/main/examples) for more examples.

## Thanks

+ [helmfile](https://github.com/helmfile/helmfile)
+ [helm-diff](https://github.com/databus23/helm-diff)
+ [helm-secrets](https://github.com/jkroepke/helm-secrets)
+ [helm-s3](https://github.com/hypnoglow/helm-s3)
+ [helm-git](https://github.com/aslafy-z/helm-git)
