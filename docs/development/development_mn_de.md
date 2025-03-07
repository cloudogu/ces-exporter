# CES-Exporter für Multinode CES
Der `ces-exporter` für Multinode CES ist eine CES-Komponente, die die Daten der Multinode CES-Instanz für die Migration in eine andere Multinode CES-Instanz bereitstellt.

## Ablauf der Migration
Der grundlegende Ablauf der Migration von einer Multinode CES-Instanz in eine andere Multinode CES-Instanz ist im folgenden Diagram beschrieben:
![Migration CES MN -> CES MN](./migration_mn-mn.png)

## API des CES-Exporters
Die API des `ces-exporter` ist in der folgenden OpenAPI-Spezifikation beschrieben: [openapi.yaml](./openapi.yaml) 

## Entwicklung im DEV CES-Cluster
Um die `ces-exporter` Komponente in einem CES-Cluster zu testen, können die folgenden Make-Targets verwendet werden:

### ApiKey
Die API des `ces-exporter` benötigt einen ApiKey. Dieser muss als Secret bereitgestellt werden.
Für die Entwicklung kann dafür das Make-Target `make apikey-secret` verwendet werden.
Dieses erstellt das Secret `ces-exporter-api` mit dem ApiKey aus der Umgebungsvariable `EXPORTER_API_KEY`.
Diese Umgebungsvariable kann auch im `.env`-File angegeben werden.

### Importer SSH Public-Key
Damit der `ces-importer` Zugriff auf die Daten des `ces-exproter` bekommt muss der SSH-PublicKey des `ces-importer` im `ces-exporter` hinterlegt werden.
Der SSH-PublicKey in den Exporter-SideCar-Container an den Dogus verwendet, um Zugriff auf die Volume-Daten des Dogus zu gewähren.

Bei der Installation des `ces-expoter` wird die ConfigMap `ces-importer-public-key` mit den PublicKey erstellt. 
Dür die Entwicklung kann die Umgebungsvariable `IMPORTER_PUBLIC_KEY` verwendet werden. 
Der Wert dieser Variable wird in die `values.yaml` getemplatet.
Diese Umgebungsvariable kann auch im `.env`-File angegeben werden.

### Installation per Helm

```bash
# Installiert das Helm-Chart von ces-exporter im Cluster (ohne die CES-Komponente)
make helm-apply
```

```bash
# Löscht das Helm-Chart von ces-exporter aus dem Cluster (ohne die CES-Komponente)
make helm-delete
```

### Installation als Component

```bash
# Installiert ces-exporter im Cluster
make component-apply
```

```bash
# Löscht ces-exporter aus dem Cluster
make component-delete
```

```bash
# Löscht ces-exporter aus dem Cluster und installiert es anschließend erneut
make component-reinstall
```

