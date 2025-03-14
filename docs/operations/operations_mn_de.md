# CES-Exporter betreiben

## Installation

`ces-exporter` kann als Komponente über den Komponenten-Operator des CES installiert werden.
Dazu muss eine entsprechende Custom-Resource (CR) für die Komponente erstellt werden.

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

Die neue yaml-Datei kann anschließend im Kubernetes-Cluster erstellt werden:

```shell
kubectl apply -f ces-exporter.yaml --namespace ecosystem
```

Der Komponenten-Operator erstellt nun die `ces-exporter`-Komponente im `ecosystem`-Namespace.

## Upgrade

Zum Upgrade muss die gewünschte Version in der Custom-Resource angegeben werden.
Dazu wird die erstellte CR yaml-Datei editiert und die gewünschte Version eingetragen.
Anschließend die editierte yaml Datei erneut auf den Cluster anwenden:

```shell
kubectl apply -f ces-exporter.yaml --namespace ecosystem
```

## Konfiguration

Die Komponente kann über das Feld `spec.valuesYamlOverwrite`. 
Alle möglichen Konfigurations-Werte sind in der [values.yaml](../../k8s/helm/values.yaml) aufgeführt.

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

### ApiKey konfigurieren
Der ApiKey dient zur Absicherung der HTTP-JSON-API des `ces-exporter`.
Der ApiKey muss in einem Secret hinterlegt werden. Der Name des Secrets und der Key müssen dann, wie [oben](#konfiguration) beschrieben, in der `spec.valuesYamlOverwrite` angegeben werden.

#### Beispiel apiKeySecret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: "ces-exporter-api"
type: Opaque
stringData:
  apiKey: "MySecretApiKey"
```

Das Secret kann dann im Cluster erstellt werden: 
```shell
kubectl apply -f api-secret.yaml --namespace ecosystem
```

Alternativ kann das Secret auch direkt mit `kubectl` erstell werden:
```shell
kubectl create secret generic ces-exporter-api --from-literal=apiKey=MySecretApiKey --namespace ecosystem
```

### SSH-PublicKey des Importers konfigurieren
Damit der `ces-importer` Zugriff auf die Daten des `ces-exproter` bekommt muss der SSH-PublicKey des `ces-importer` im `ces-exporter` hinterlegt werden.
Der SSH-PublicKey in den Exporter-SideCar-Container an den Dogus verwendet, um Zugriff auf die Volume-Daten des Dogus zu gewähren.

Der PublicKey muss bei der Installation des `ces-exporter`, wie [oben](#konfiguration) beschrieben, in der `spec.valuesYamlOverwrite` angegeben werden.