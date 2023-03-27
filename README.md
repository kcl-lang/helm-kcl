# Helm KCL Plugin

Helm KCL Plugin

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

Make sure you do not have a version of `helm-kcl` installed. You can remove it by running `helm plugin uninstall kcl`.

### Installation Steps

The first step is to download the repository and enter the directory. You can do this via git clone or downloading and extracting the release. If you clone via git, remember to check out the latest tag for the latest release.

Next, depending on which helm version you have, install the plugin into helm.

#### Helm 2

make install

#### Helm 3

make install/helm3
