## Operate CES exporter

## Installation

`ces-exporter` can be installed as a component via the CES component operator.
To do this, a corresponding custom resource (CR) must be created for the component.

```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: ces-exporter
  labels:
    app: ces
spec:
  name: ces-exporter
  namespace: k8s
  version: 0.0.1
```

The new yaml file can then be created in the Kubernetes cluster:

```shell
kubectl apply -f ces-exporter.yaml --namespace ecosystem
```

The component operator now creates the `ces-exporter` component in the `ecosystem` namespace.

## Upgrade

To upgrade, the desired version must be specified in the custom resource.
To do this, the CR yaml file created is edited and the desired version is entered.
Then reapply the edited yaml file to the cluster:

```shell
kubectl apply -f ces-exporter.yaml --namespace ecosystem
```

## Configuration

The component can be configured via the `spec.valuesYamlOverwrite` field.
All possible configuration values are listed in the [values.yaml](../../k8s/helm/values.yaml).

### Beispiel valuesYamlOverwrite:
```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: ces-exporter
  labels:
    app: ces
spec:
  name: ces-exporter
  namespace: k8s
  version: 0.0.1
  valuesYamlOverwrite: |
    env:
      logLevel: info
      basePath: "/ces-exporter"
    apiKey:
      secretName: "ces-exporter-api"
      secretDataKey: "apiKey"
    publicKey:
      configMapName: "ces-importer-public-key"
      configMapDataKey: "publicKey"
      data: |-
        ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC9gUMHJITGQbeVqpSvxb5D8eaB3ZCrpnIBeQr73LKBpWiAdmEOBrVNfpWmFARrkHSFP51uLNHUxjASYHDcprqF1LbgMCZD6zvJjMDwEzYaumK/Z3d/pYq3em4O5w2E+ktQn7vGfKqmBqrZoo5ghPArbb1hpqY5cjntb1X71BYYu0vk9HhpVsxb7dRFZZVMmOK1DsD8509o441Tl89oPgf+1KIBpGkoPdUenGqYPDVLaiIHEliaHCVUWXel3XZu3kITCEsrGoTJkOtnGAcSW/4rmXO7KwKVXHhnCECCGpktzYq+O27kp7BPN87VDb9/f9WjReXlj55/GgDU1k71N0a26yIuoSaHckkAeZjhyylVuHOkMrQku2tti7czIUhVtK7VBuajwckGLFfYdlOd3P/xvpKAI9s5R8hBHMuczMcpGWTvSTFBWPudrL4U7fuJgyfuScFL4UOPOl4Lj9DzV18so6zwTrWF6YRlwHuDDuUdMQsuOPr+QcvPXH2Y7wEJtUWBMeA63Q1mI3W/VjxV5TFqA+mkwCqf6vg6xlqUtn9y2Mc8W+2+StSmnlN5iOIQ12sCDZT3faJavalqe3iPWRr4huf1mW2mwlCDZp/Lc0swnlSZHG0nvyGWhsS9ZjpuHVH38Sgre9SLFpjYqWjjfGWdU5DU2pgJs5iZ/LiTSX1jhQ== ces-importer@server01
```

### Configure ApiKey
The ApiKey is used to secure the HTTP JSON API of the `ces-exporter`.
The ApiKey must be stored in a secret. The name of the secret and the key must then be specified in the `spec.valuesYamlOverwrite` as described [above](#configuration).

#### Example apiKeySecret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: "ces-exporter-api"
type: Opaque
stringData:
  apiKey: "MySecretApiKey"
```

The secret can then be created in the cluster:
```shell
kubectl apply -f api-secret.yaml --namespace ecosystem
```

Alternatively, the secret can also be created directly with ``kubectl``:
```shell
kubectl create secret generic ces-exporter-api --from-literal=apiKey=MySecretApiKey --namespace ecosystem
```

### Configure SSH-PublicKey of the importer
To give the `ces-importer` access to the data of the `ces-exproter`, the SSH-PublicKey of the `ces-importer` must be stored in the `ces-exporter`.
The SSH-PublicKey in the Exporter-SideCar-Container to the Dogus is used to grant access to the volume data of the Dogus.

The PublicKey must be specified in the `spec.valuesYamlOverwrite` during the installation of the `ces-exporter`, as described [above](#configuration).
