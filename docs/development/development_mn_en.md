# CES exporter for Multinode CES
The `ces-exporter` for Multinode CES is a CES component that provides the data of the Multinode CES instance for migration to another Multinode CES instance.

## Migration procedure
The basic migration process from one multinode CES instance to another multinode CES instance is described in the following diagram:
[Migration CES MN -> CES MN](https://docs.cloudogu.com/en/usermanual/automatic_migration/reference/migration_flow/#migration-ces-mn---ces-mn)

## API of the CES exporter
The API of the `ces-exporter` is described in the following OpenAPI specification: [openapi.yaml](./openapi.yaml)

## Development in the DEV CES cluster
To test the `ces-exporter` component in a CES cluster, the following make targets can be used:

### ApiKey
The API of the `ces-exporter` requires an ApiKey. This must be provided as a secret.
The Make target `make apikey-secret` can be used for development.
This creates the secret `ces-exporter-api` with the ApiKey from the environment variable `EXPORTER_API_KEY`.
This environment variable can also be specified in the `.env` file.

### Importer SSH public key
To give the `ces-importer` access to the data of the `ces-exporter`, the SSH public key of the `ces-importer` must be stored in the `ces-exporter`.
The SSH-PublicKey in the Exporter-SideCar-Container to the Dogus is used to grant access to the volume data of the Dogus.

When installing the `ces-exporter`, the ConfigMap `ces-importer-public-key` is created with the PublicKey.
The environment variable `IMPORTER_PUBLIC_KEY` can be used for development.
The value of this variable is templated in the `values.yaml`.
This environment variable can also be specified in the `.env` file.

### Dependencies

#### Exposition-CRD
To expose the exporter port, the component creates an Exposition-CR in the source system.
This requires the corresponding CustomResourceDefinition to be installed.

```bash
# Installs the Exposition-CRD
 helm install k8s-exposition-crd oci://registry.cloudogu.com/k8s/k8s-exposition-crd --version 1.0.0 --namespace ecosystem
```

or via component-yaml:
```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: k8s-exposition-crd
spec:
  name: k8s-exposition-crd
  namespace: k8s
  version: 1.0.0
```

#### K8s-Service-Discovery
Ensure that the K8s-Service-Discovery is installed in a version >= 6.1.0 and configured for exposition-CRs, so that the ssh port of the exporter will get needed resources.

```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: k8s-service-discovery
  namespace: ecosystem
spec:
  name: k8s-service-discovery
  namespace: k8s
  version: 6.1.0
  valuesYamlOverwrite: |-
    exposition:
      discoverExpositionCR: true
```

#### K8s-Ces-Gateway
The K8s-Ces-Gateway is required to expose the exporter ssh port. This is statically configured in the version `3.3.3` and above.

```yaml
apiVersion: k8s.cloudogu.com/v1
kind: Component
metadata:
  name: k8s-ces-gateway
  namespace: ecosystem
spec:
  name: k8s-ces-gateway
  namespace: k8s
  version: 3.3.3
```

### Installation via helm

```bash
# Installs the Helm chart from ces-exporter in the cluster (without the CES component)
make helm-apply
```

```bash
# Deletes the helmet chart of ces-exporter from the cluster (without the CES component)
make helm-delete
```

### Installation as component

```bash
# Installs ces-exporter in the cluster
make component-apply
```

```bash
# Deletes ces-exporter from the cluster
make component-delete
```

```bash
# Deletes ces-exporter from the cluster and then reinstalls it
make component-reinstall
```
