# CES-Exporter für Classic CES
Der `ces-exporter` für Classic CES ist eine CES-Komponente, die die Daten der Classic CES-Instanz für die Migration in eine Multinode CES-Instanz bereitstellt.

## Bauen des Debian-Paketes

Das Debian Paket enthält das Docker-Image der jeweiligen gebauten Version des ces-exporters als Archiv.
Bei der Installation wird das Archiv als Docker-Image importiert und gestartet. Das geschieht über einen Systemd-Service.

```bash
# Erstellt das Debian Paket
make debian
```

Wichtig: Das Paket darf **nicht** mit `make package` gebaut werden, da ansonsten das Binary des ces-exporters im Hostsystem abgelegt wird, statt nur im Container.
