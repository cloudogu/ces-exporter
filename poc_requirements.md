<!-- Anforderungen und Abnahmekriterien für den V3-Exporter-PoC. -->

# Requirements PoC

Basis ist ADR 0067 im Repository `k8s-ecosystem-architecture`.
Kontext ist die automatische Migration von V3-Dogus. Der PoC untersucht die
Bereitstellung ihrer Daten über Exporter-Sidecars; die Migration selbst wird
nicht getestet.

## Ziel und Testobjekte

Der PoC soll nachweisen, dass sich die migrationsrelevanten PVCs von V3-Dogus
über den in dem ADR beschriebenen Prozess generisch in einen Exporter-Sidecar
einbinden und dort lesen lassen. Aktivierung und Deaktivierung müssen
automatisch funktionieren.

Der Nachweis erfolgt mit Nexus `3.96.4-2` als V3-Dogu und dessen im Helm-Chart
mitgelieferter PostgreSQL-Instanz in einem Test-Namespace. Beide Workloads
gehören zur selben V3-Dogu-CR; eine separate PostgreSQL-Dogu-CR ist nicht nötig.
Ein erfolgreicher
manueller Happy-Path-Durchlauf genügt als Abnahme. Dieser belegt den Mechanismus
an den Testobjekten, nicht sämtliche Volume- und Workload-Konstellationen.

## Aktivierung und Deaktivierung

- Es gibt zwei getrennte ConfigMaps: eine zum Triggern des PoC über die
  Aktivierung des Exportmodus und eine, anhand derer der Admission Webhook
  entscheidet, ob ein Pod mit Exporter-Sidecar gestartet werden muss.
- Das Anlegen der ConfigMap `ces-exporter-export-mode` aktiviert den Exportmodus
  für alle V3-Dogus im Namespace dieser ConfigMap. Für den Trigger genügt ihre
  Existenz; zusätzliche Datenfelder sind nicht erforderlich.
- Beim Einschalten des Exportmodus setzt der Exporter (Export-Controller)
  zuerst den Exportmodus-Wert in der separaten ConfigMap für den Admission
  Webhook auf `"true"`. Anschließend veranlasst er die Neuerzeugung aller für
  die Migration betroffenen Pods im Namespace. Der Webhook muss den aktivierten
  Wert bei der Erstellung dieser Pods bereits berücksichtigen.
- Die ConfigMap startet keinen vollständigen automatischen Testlauf. Der
  Exportmodus bleibt aktiv, solange die ConfigMap existiert; die Prüfungen
  erfolgen manuell.
- Das Löschen der Trigger-ConfigMap löst die Deaktivierung aus. Der Exporter
  setzt zuerst den Exportmodus-Wert in der Webhook-ConfigMap auf `"false"` und
  veranlasst anschließend die Neuerzeugung der betroffenen Pods ohne Sidecar.

## Pod-Ersetzung und Sidecar-Injektion

- Der Export-Controller veranlasst bei Aktivierung und Deaktivierung die
  Ersetzung der betroffenen Pods. Manuelle Pod-Eingriffe sind dafür nicht nötig.
- Ein Mutating Admission Webhook ergänzt bei der Erstellung eines Pods im
  Exportmodus den Exporter-Sidecar, die erforderlichen Volume-Mounts und das
  benötigte Schlüsselmaterial gemäß ADR 0067.
- Der Webhook verwendet dafür die separate ConfigMap als Entscheidungsgrundlage
  für die Sidecar-Injektion. Diese muss den jeweils aktuellen Exportmodus der
  betroffenen Pods abbilden, auch nach der Deaktivierung.
- Die Helm-Workload-Templates werden dafür nicht verändert.
- Die Pod-Ersetzung folgt der ADR: nacheinander, unter Berücksichtigung der
  Bereitschaft von Anwendung und Exporter. Downtime ist akzeptiert.
- Nach der Deaktivierung laufen die Anwendungen wieder bereit, ohne
  Exporter-Sidecar und ohne die ausschließlich dafür ergänzten Mounts und
  Schlüssel. Die Anwendungsdaten bleiben erhalten.

## Volumes und Lesenachweis

- Die Volume-Zuordnung wird generisch aus den Kubernetes-Ressourcen ermittelt.
  Nexus- oder PostgreSQL-spezifische Volume-Namen und Pfade werden nicht fest
  eingebaut.
- Berücksichtigt werden alle migrationsrelevanten PVCs gemäß ADR 0067;
  insbesondere bleiben `existingClaim`-PVCs ausgeschlossen.
- Andere Volume-Typen wie ConfigMaps, Secrets, `projected` und `emptyDir`
  gehören nicht zum Volume-Nachweis.
- Der Exporter erhält ausschließlich lesenden Zugriff auf die betreffenden
  PVCs. Die bisherigen Mount-Rechte der Anwendungen bleiben erhalten.
- Als einfacher Nachweis wird pro relevantem PVC eine vorhandene Datei aus
  dem Exporter gelesen. Ein zusätzlicher Inhaltsvergleich mit dem
  Anwendungscontainer ist nicht erforderlich.

## Abnahmeszenario

1. Nexus `3.96.4-2` als V3-Dogu mit dem mitgelieferten PostgreSQL im
   Test-Namespace bereitstellen.
   Die tatsächlichen PVCs und Mounts ermitteln und geeignete Dateien für den
   Lesenachweis auswählen.
2. Die ConfigMap `ces-exporter-export-mode` im Test-Namespace anlegen.
3. Prüfen, dass der Exporter zuerst den Exportmodus-Wert in der separaten
   ConfigMap für den Admission Webhook auf `"true"` setzt und anschließend alle
   für die Migration betroffenen Pods automatisch neu erzeugen lässt. Prüfen,
   dass der Webhook diesen Wert zur Sidecar-Injektion verwendet und die
   Anwendungen sowie ihre Exporter bereit sind.
4. Für jedes relevante PVC eine Datei aus dem jeweiligen Exporter lesen.
5. Die ConfigMap löschen.
6. Prüfen, dass der Exporter zuerst den Exportmodus-Wert in der
   Webhook-ConfigMap auf `"false"` setzt und anschließend die Pods automatisch
   erneut ersetzt werden. Die Anwendungen müssen ohne Exporter sowie dessen
   zusätzliche Mounts und Schlüssel bereit sein.
7. Prüfen, dass die zuvor betrachteten Anwendungsdaten weiterhin vorhanden sind.

## Dokumentation und Abgrenzung

Alle für den erfolgreichen Durchlauf nötigen Vorbereitungen, Installations-
und Konfigurationsschritte, Befehle und manuellen Prüfungen werden dokumentiert.
Das schließt die Einrichtung und Registrierung des Webhooks einschließlich
seiner Zertifikate ein. Manuelle Vorbereitung ist zulässig; ein automatisierter
Test oder eine automatisierte Installation ist nicht erforderlich.

Nicht Teil der PoC-Abnahme sind:

- Die eigentliche Migration oder ein SSH-Datentransfer.
- Fehlerfalltests und der Nachweis vollständiger Fehlerbehandlung.
- Die automatische Ermittlung und Aktivierung von Dogu-Abhängigkeiten; das
  Signal gilt bereits für alle V3-Dogus im Namespace.
- Nachweise für weitere Dogus, mehrere Replicas oder andere
  Workload-Konstellationen über den konkreten Testaufbau hinaus.

Die tatsächlichen PVCs, Mounts und das Workload-Layout der Testobjekte sind vor
dem Durchlauf technisch zu ermitteln. Der genaue Mechanismus der Pod-Ersetzung
und die Mount-Pfade des Exporters sind Implementierungsentscheidungen innerhalb
der Vorgaben von ADR 0067.

Die Umsetzung und der manuelle Prüfablauf sind in
[V3-Exporter-PoC](docs/development/poc_v3_exporter_de.md) dokumentiert.
