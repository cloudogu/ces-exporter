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
