# CES exporter for Classic CES
The `ces-exporter` for Classic CES is a CES component that provides the data of the Classic CES instance for migration to a multinode CES instance.

## Building the Debian package

The Debian package contains the Docker image of the respective built version of the ces-exporter as an archive.
During installation, the archive is imported as a Docker image and started. This is done via a systemd service.

```bash
# Creates the Debian package
make debian
```

Important: The package must **not** be built with `make package`, otherwise the binary of the ces-exporter will be stored in the host system instead of just in the container.

## Configuration
In order for the ces-exporter to be used correctly, an SSH public key and an API key must be stored.

The API key can be freely selected and enables access to the application's interfaces. The key entered here
must also be entered correctly in the importer for access to work.
To set the key for the exporter, this must simply be stored in etcd. The key responsible for this in
etcd is `/config/ces-exporter/authentication/api_key`.

The SSH key must be a valid public key and match the private key stored in the importer. It must have been generated
with a corresponding SSH command (ssh-keygen). The key cannot have a password. In order to store this correctly in the exporter,
must be saved in etcd. The key responsible for this in etcd is
`/config/ces-exporter/authentication/public_key`.

## Using the api key

Requests to the exporter API need to be using the `X-CES-EXPORTER-API-KEY` header.