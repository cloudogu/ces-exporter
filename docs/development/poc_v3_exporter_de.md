<!-- Installationsanleitung, manueller Prüfablauf und dokumentierte Abnahme des V3-Exporter-PoC. -->

# V3-Exporter-PoC

Eine ausführliche Erklärung aller beteiligten Dateien und ihres Zusammenspiels
steht in [Die Dateien des V3-Exporter-PoC erklärt](poc_v3_dateien_de.md).

Der PoC ergänzt im Multinode-CES einen optionalen Controller und einen TLS
Mutating Admission Webhook. Er stellt migrationsrelevante PVCs über einen
Exporter-Sidecar lesend bereit. Die eigentliche Migration und ein SSH-Transfer
sind nicht Teil der Abnahme.

## Voraussetzungen

- Kubernetes mit der Dogu-API `k8s.cloudogu.com/v3beta1`, V3-fähigem Dogu-Operator
  und dessen Helm-/Flux-Anbindung sowie cert-manager einschließlich CA-Injector.
- Zugriff auf die V3-Nexus-Registry und das Sidecar-Image
  `registry.cloudogu.com/k8s/rsync-sidecar:1.1.2`.
- `kubectl`, Helm, Go und ein Container-Build-Werkzeug.
- Ein Test-Namespace, in dem der Dogu-Operator Dogus verwaltet. Der hier geprüfte
  Operator ist auf `ecosystem` beschränkt.

Für den vorhandenen k3d-Cluster:

```bash
export KUBECONFIG=/home/coder/.kube/ecosystem-v3.k3ces.localdomain
export NAMESPACE=ecosystem
kubectl get nodes
kubectl get certificates -n "$NAMESPACE"
```

Das vorhandene Blueprint bleibt pausiert. Plattform-Komponenten werden durch den
PoC nicht ausgewählt. Die Trigger-ConfigMap gilt für alle V3-Dogus im Namespace;
für diesen Test wird nur Nexus installiert. PostgreSQL ist ein eigener StatefulSet
im Nexus-Chart und gehört zur selben V3-Dogu-CR.

## Testobjekte installieren

```bash
kubectl apply -f k8s/poc/dogus.yaml
kubectl wait --for=condition=ready dogu/nexus -n "$NAMESPACE" --timeout=10m
kubectl get statefulsets,pods,pvc -n "$NAMESPACE"
```

Das Manifest installiert Nexus `3.96.4-2` mit `spec.doguApiVersion: v3`. Dessen
PostgreSQL wird vom Chart bereitgestellt; es wird keine separate PostgreSQL-CR
angelegt. Bei Registry- oder Installationsfehlern zuerst deren Ursache beheben.

## Exporter bauen und installieren

Das vorhandene Dockerfile kann mit `--target multinode` gebaut werden. Für einen
schnellen lokalen k3d-Test mit dem lokal installierten Go-Compiler:

```bash
POC_BUILD_DIR=$(mktemp -d)
CGO_ENABLED=0 go build -o "$POC_BUILD_DIR/ces-exporter" .
cat > "$POC_BUILD_DIR/Dockerfile" <<'DOCKERFILE'
FROM gcr.io/distroless/static:nonroot
ENV MODE=multinode
COPY ces-exporter /ces-exporter
USER 65532:65532
ENTRYPOINT ["/ces-exporter"]
DOCKERFILE
docker build -t docker.io/local/ces-exporter:poc "$POC_BUILD_DIR"
k3d image import docker.io/local/ces-exporter:poc -c ecosystem-v3
```

Docker und k3d benötigen auf diesem Entwicklungsrechner `sudo`. Bei einem
weiteren Build einen neuen Image-Tag verwenden, damit Kubernetes das neue Image
lädt.

Einen Testschlüssel und den API-Key lokal erzeugen; private Schlüssel bleiben
außerhalb des Repositorys:

```bash
POC_KEYS_DIR=$(mktemp -d)
ssh-keygen -q -t ed25519 -N '' -f "$POC_KEYS_DIR/importer"
openssl rand -hex 24 > "$POC_KEYS_DIR/api-key"
kubectl create secret generic ces-exporter-api -n "$NAMESPACE" \
  --from-file=apiKey="$POC_KEYS_DIR/api-key"
helm upgrade --install ces-exporter k8s/helm -n "$NAMESPACE" \
  --set poc.enabled=true \
  --set image.registry=docker.io \
  --set image.repository=local/ces-exporter \
  --set image.tag=poc \
  --set-file publicKey.data="$POC_KEYS_DIR/importer.pub"
kubectl wait --for=condition=Ready certificate/ces-exporter-poc \
  -n "$NAMESPACE" --timeout=2m
kubectl rollout status deployment/ces-exporter -n "$NAMESPACE" --timeout=2m
```

Bei vorhandenen Schlüsseln/API-Key-Secrets diese wiederverwenden. Der öffentliche
Schlüssel liegt in `ces-importer-public-key` unter `publicKey`. Die Helm-Werte
`publicKey.configMapName` und `publicKey.configMapDataKey` konfigurieren zugleich
den Webhook. `poc.sidecarImage` überschreibt das Sidecar-Image.

Helm erstellt einen SelfSigned-Issuer, ein Serving-Certificate mit den DNS-Namen
des Webhook-Service und die MutatingWebhookConfiguration. cert-manager stellt
das Secret `ces-exporter-poc-tls` bereit und injiziert die CA. Der Exporter lädt
die gemounteten Zertifikate für TLS-Verbindungen neu. Der Controller wartet auf
CA, Webhook-Registrierung und einen bereiten Service-Endpunkt. Die
Webhook-Registrierung gilt ausschließlich für Pod-`CREATE` im Release-Namespace.
Sie verwendet `failurePolicy: Ignore`; Injektionsfehler werden protokolliert und
führen nicht zu einer erfolgreichen Export-Abnahme.

## Ressourcen vorab ermitteln

```bash
kubectl get pods -n "$NAMESPACE" \
  -o custom-columns=NAME:.metadata.name,UID:.metadata.uid
kubectl get statefulsets -n "$NAMESPACE" \
  -l k8s.cloudogu.com/dogu.name=nexus -o yaml
kubectl get pod nexus-nexus-0 -n "$NAMESPACE" -o yaml
kubectl get pod nexus-nexus-postgresql-0 -n "$NAMESPACE" -o yaml
```

Der Webhook folgt den Workload-OwnerReferences und ordnet den Workload über
`k8s.cloudogu.com/dogu.name` einer V3-Dogu-CR zu. PVCs werden nur berücksichtigt,
wenn sie im installierten Helm-Release-Manifest stehen oder aus einem
StatefulSet-Claim-Template des Workloads stammen. Externe `existingClaim`-PVCs
werden dadurch ausgeschlossen. Block-Volumes sind für diesen PoC nicht
unterstützt. Die bestehenden Anwendungs-Mounts bleiben unverändert.
Der Sidecar mountet das gesamte PVC; bei Anwendungs-Mounts mit `subPath` muss
dieser deshalb dem relativen Dateipfad im Sidecar vorangestellt werden.

Im geprüften Chart ergibt sich folgende Zuordnung. Diese Namen gehören nur zur
Anleitung und sind nicht im Controller oder Webhook fest eingebaut:

| Pod | PVC | Anwendungspfad | Datei für den Lesenachweis |
| --- | --- | --- | --- |
| `nexus-nexus-0` | `nexus-data-nexus-nexus-0` | `/var/lib/nexus` (`subPath: data`) | `data/etc/nexus.properties` |
| `nexus-nexus-postgresql-0` | `postgresql-data-nexus-nexus-postgresql-0` | `/var/lib/postgresql/data` | `pgdata/PG_VERSION` |

## Aktivieren, lesen und deaktivieren

```bash
kubectl create configmap ces-exporter-export-mode -n "$NAMESPACE"
kubectl get configmap ces-exporter-webhook-config -n "$NAMESPACE" -o yaml
kubectl get pods -n "$NAMESPACE" -w
```

Zuerst setzt der Exporter `data.exportMode` in
`ces-exporter-webhook-config` auf `"true"`. Dann ersetzt er die betroffenen Pods
nacheinander und wartet jeweils auf `Ready`, einschließlich des Exporters.
Die UID-Vorbedingung beim Löschen schützt zwischenzeitliche Ersatz-Pods.
Neue Pods erhalten den Container `ces-exporter-sidecar`, schreibgeschützte
PVC-Mounts unter `/export/<pvc-name>` und den öffentlichen SSH-Schlüssel.
Die Workload-Templates werden nicht verändert. Pod-UIDs müssen sich gegenüber
dem Ausgangszustand geändert haben.

Auf `data.phase: ready` warten und dann aus jedem Exporter eine Datei lesen.
Ein Hash liest die Datei vollständig, ohne Konfigurationsinhalte auszugeben:

```bash
kubectl exec -n "$NAMESPACE" nexus-nexus-0 -c ces-exporter-sidecar -- \
  sha256sum /export/nexus-data-nexus-nexus-0/data/etc/nexus.properties
kubectl exec -n "$NAMESPACE" nexus-nexus-postgresql-0 -c ces-exporter-sidecar -- \
  cat /export/postgresql-data-nexus-nexus-postgresql-0/pgdata/PG_VERSION
kubectl get pod nexus-nexus-0 -n "$NAMESPACE" -o yaml
kubectl get pod nexus-nexus-postgresql-0 -n "$NAMESPACE" -o yaml
kubectl delete configmap ces-exporter-export-mode -n "$NAMESPACE"
kubectl get configmap ces-exporter-webhook-config -n "$NAMESPACE" -o yaml
kubectl get pods -n "$NAMESPACE" -w
```

Die Sidecar-Mounts müssen `readOnly: true` enthalten. Bei Deaktivierung wird zuerst
`data.exportMode: "false"` gesetzt. Nach der erneuten sequenziellen Ersetzung
müssen beide Anwendungen bereit sein, ohne Sidecar, Schlüsselvolume und dessen
zusätzliche Mounts. Anschließend die Dateien in den Anwendungen erneut lesen:

```bash
kubectl exec -n "$NAMESPACE" nexus-nexus-0 -c nexus -- \
  sha256sum /var/lib/nexus/etc/nexus.properties
kubectl exec -n "$NAMESPACE" nexus-nexus-postgresql-0 -c postgresql -- \
  cat /var/lib/postgresql/data/pgdata/PG_VERSION
```

Pro Pod wartet der Controller maximal zehn Minuten. Bei einem Fehler setzt er
`data.phase: failed` und `data.error` und erzeugt ein Warning-Event. Derselbe
Trigger wird auch nach Exporter-Neustarts nicht erneut abgearbeitet; es gibt
keine dauerhafte Neustartschleife. Nach Beheben der Ursache den Trigger löschen
und neu anlegen. Bei fehlgeschlagener Deaktivierung kann nach Ursachenbehebung
das Feld `phase` der Webhook-ConfigMap manuell entfernt werden.

```bash
kubectl logs deployment/ces-exporter -n "$NAMESPACE"
kubectl get events -n "$NAMESPACE" --sort-by=.lastTimestamp
```

Nach der Abnahme bleibt der Exportmodus deaktiviert. Test-Dogus und Exporter
können für weitere Untersuchungen installiert bleiben. Zum Entfernen zuerst die
Trigger-ConfigMap löschen und die erfolgreiche Deaktivierung abwarten, dann den
Exporter per Helm deinstallieren. Das Löschen der Test-Dogu-CR ist ein separater
Schritt; PVCs und Anwendungsdaten nicht pauschal löschen.

## Lokale Prüfungen

```bash
go test ./exportpoc -race
go test ./...
helm lint k8s/helm
helm lint k8s/helm --set poc.enabled=true
helm template ces-exporter k8s/helm -n ecosystem --set poc.enabled=true
```

Der vorhandene `Test_main` benötigt exklusiven Zugriff auf Port 8080. Auf diesem
Entwicklungsrechner belegt ein fremder Prozess `127.0.0.2:8080`; deshalb wurden
die übrigen Tests mit `-skip '^Test_main$'` und dieser Test separat in einem
Netzwerk-Namespace ausgeführt. Der fremde Prozess wurde nicht verändert.

## Abnahme am 25.11.2026

Der manuelle Durchlauf im k3d-Cluster `ecosystem-v3`, Namespace `ecosystem`,
war erfolgreich. Verwendet wurden Nexus `3.96.4-2` als V3-Dogu und dessen
mitgelieferter PostgreSQL-StatefulSet.

| Pod | UID vor Aktivierung | UID mit Sidecar | UID nach Deaktivierung |
| --- | --- | --- | --- |
| `nexus-nexus-0` | `89ddf2ac-e124-4e7e-9241-b9dabf99dab6` | `5a92e70f-02ca-493f-a850-9bf997e60f48` | `b3e5eaef-a038-434f-b595-6fd59ec59bba` |
| `nexus-nexus-postgresql-0` | `364bf6b4-b68b-4d65-9faa-156d620e4c5a` | `c9f5201d-6b9c-499a-a5dc-bfd79ce0e548` | `94b7adb3-53de-41dc-afab-fd7fc0ff095c` |

In beiden Sidecars war jeder ergänzte Volume-Mount schreibgeschützt. Die Dateien
wurden aus beiden PVCs vollständig gelesen. Ihre SHA-256-Werte waren vor der
Aktivierung, im Exporter und nach der Deaktivierung identisch:

| PVC und Datei | SHA-256 |
| --- | --- |
| `nexus-data-nexus-nexus-0/data/etc/nexus.properties` | `429bf66ede0e0007c4f895f8306b0db60ce6438d917192c5d6dedbc1334157cb` |
| `postgresql-data-nexus-nexus-postgresql-0/pgdata/PG_VERSION` | `9a92adbc0cee38ef658c71ce1b1bf8c65668f166bfb213644c895ccb1ad07a25` |

Die JSON-Hashes der StatefulSet-Pod-Templates waren in allen drei Phasen
identisch; der Webhook hat ausschließlich neue Pods ergänzt. Nach der
Deaktivierung waren Nexus `2/2` und PostgreSQL `1/1` bereit. Beide Pods hatten
weder den Exporter-Container noch dessen Schlüsselvolume und zusätzliche Mounts.
Die Trigger-ConfigMap war gelöscht; die Webhook-ConfigMap enthielt
`exportMode: "false"`, `phase: ready` und `transition: inactive`.

Anschließend wurde der Exporter auf das Image
`docker.io/local/ces-exporter:poc-verified` aktualisiert. Nach diesem Neustart
blieben die Anwendungs-Pod-UIDs, Daten-Hashes und Workload-Templates unverändert;
der Exportmodus blieb deaktiviert und bereit.

Erfolgreiche lokale Prüfungen: vollständiges `go test ./... -count=1` im
isolierten Netzwerk, gezielte Go-Tests einschließlich Race-Detector, Helm-Lint
mit und ohne PoC sowie serverseitige Dry-Run-Validierung der gerenderten Manifeste.
