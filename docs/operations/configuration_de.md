# Konfiguration des API-Keys

## Konfiguration im Classic-CES
Damit der ces-exporter korrekt verwendet werden kann, müssen ein SSH-Public-Key und ein Api-Key hinterlegt werden.

Der Api-Key kann frei gewählt werden und ermöglicht den Zugriff auf die Schnittstellen der Anwendung. Der Key, der 
hier angegeben wird, muss ebenfalls im Importer korrekt angegeben sein, damit der Zugriff funktioniert.
Um den Key für den Exporter zu setzen, muss dieser einfach im etcd hinterlegt werden. Der dafür zuständige Key im 
etcd ist `/config/ces-exporter/authentication/api_key`.

Der SSH-Key muss ein gültiger Public-Key sein und zu dem private-key passen, der im Importer hinterlegt ist. Er muss 
mit einem entsprechenden SSH-Befehl generiert worden sein. Um diesen nun korrekt im Exporter zu hinterlegen, muss 
dieser im etcd abgespeichert werden. Der dafür zuständige Key im etcd ist 
`/config/ces-exporter/authentication/public_key`.

## Nutzung des API-Keys

Bei Anfragen an die API des Exporters muss der API-Schlüssel als Header `X-CES-EXPORTER-API-KEY` angegeben werden.
