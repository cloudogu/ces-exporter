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

## Konfiguration
Damit der ces-exporter korrekt verwendet werden kann, müssen ein SSH-Public-Key und ein Api-Key hinterlegt werden.

Der Api-Key kann frei gewählt werden und ermöglicht den Zugriff auf die Schnittstellen der Anwendung. Der Key, der
hier angegeben wird, muss ebenfalls im Importer korrekt angegeben sein, damit der Zugriff funktioniert.
Um den Key für den Exporter zu setzen, muss dieser einfach im etcd hinterlegt werden. Der dafür zuständige Key im
etcd ist `/config/ces-exporter/authentication/api_key`.

Der SSH-Key muss ein gültiger Public-Key sein und zu dem private-key passen, der im Importer hinterlegt ist. Er muss
mit einem entsprechenden SSH-Befehl (ssh-keygen generiert worden sein. Der Key muss ohne Passwort angelegt werden. Um diesen nun korrekt im Exporter zu hinterlegen, muss
dieser im etcd abgespeichert werden. Der dafür zuständige Key im etcd ist
`/config/ces-exporter/authentication/public_key`.

## Nutzung des API-Keys

Bei Anfragen an die API des Exporters muss der API-Schlüssel als Header `X-CES-EXPORTER-API-KEY` angegeben werden.
