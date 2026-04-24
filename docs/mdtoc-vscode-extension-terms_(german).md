# Begriffe zur möglichen VS-Code-Extension für `mdtoc`

Diese Datei erläutert die wichtigsten Begriffe und Abkürzungen aus der bisherigen Einschätzung zur Frage, wie `mdtoc` zukünftig als VS-Code-Extension verfügbar gemacht werden kann, während die bestehende Go-CLI erhalten bleibt.

## MVP

**MVP** steht für **Minimum Viable Product**, auf Deutsch etwa: **kleinste sinnvoll nutzbare Erstversion**.

Für deine VS-Code-Extension hieße das:

> Nicht sofort alle denkbaren Features bauen, sondern zuerst eine Version, die den Hauptnutzen liefert und technisch sauber funktioniert.

Ein sinnvolles MVP für `mdtoc` wäre zum Beispiel:

```text
- Aktive Markdown-Datei öffnen
- Befehl "mdtoc: Generate ToC" ausführen
- bestehende Go-CLI im Hintergrund aufrufen
- Ergebnis zurück in den Editor schreiben
- Fehler verständlich anzeigen
```

Noch **nicht** im MVP enthalten sein müssten:

```text
- automatische Ausführung beim Speichern
- Marketplace-Veröffentlichung für alle Plattformen
- Web-Version für vscode.dev
- Live-Diagnostics
- grafische Einstellungen
- Preview-Modus
- Workspace-weite Verarbeitung vieler Dateien
```

Der Vorteil eines MVP: Du erkennst früh, ob die Architektur passt, ohne dich schon in Packaging, Marketplace, Plattform-Binaries und Sonderfällen zu verlieren.

---

## VS Code

**VS Code** ist die Desktop-Anwendung **Visual Studio Code**.

Wichtig ist die Unterscheidung zwischen:

```text
VS Code Desktop
```

und

```text
VS Code im Browser
```

Für dein Go-basiertes Tool ist diese Unterscheidung zentral.

In **VS Code Desktop** kann eine Extension normalerweise Node.js-Funktionen verwenden und dadurch auch ein externes Programm wie dein Go-Binary `mdtoc` starten.

In **VS Code im Browser** geht das nicht ohne Weiteres.

---

## `vscode.dev`

`vscode.dev` ist die Browser-Version von Visual Studio Code.

Man öffnet sie im Browser, ohne lokal VS Code zu installieren. Sie läuft also ungefähr so:

```text
Browser
└─ vscode.dev
   └─ VS-Code-Oberfläche
```

Der wichtige Punkt:

> Eine Extension auf `vscode.dev` läuft in einer Browser-Sandbox.

Das bedeutet: Sie kann nicht einfach ein lokales ausführbares Programm wie `mdtoc.exe`, `mdtoc` oder ein Go-Binary starten.

Für `mdtoc` heißt das:

```text
Go-CLI direkt ausführen
→ in VS Code Desktop möglich

Go-CLI direkt ausführen
→ in vscode.dev nicht möglich
```

Wenn du `mdtoc` auch in `vscode.dev` verfügbar machen wolltest, bräuchtest du wahrscheinlich eine andere technische Strategie:

```text
- Kernlogik nach TypeScript portieren
oder
- Go-Code nach WebAssembly kompilieren
oder
- Remote-/Server-Komponente verwenden
```

Für eine erste Version würde ich `vscode.dev` daher bewusst **nicht** unterstützen.

---

## `github.dev`

`github.dev` ist eine VS-Code-ähnliche Browser-Editor-Umgebung direkt für GitHub-Repositories.

Wenn du in einem GitHub-Repository bist, kannst du typischerweise `.` drücken oder die URL entsprechend ändern, um das Repository im Browser-Editor zu öffnen.

Technisch ist `github.dev` ähnlich relevant wie `vscode.dev`:

```text
github.dev läuft im Browser
→ keine normale Ausführung lokaler Go-Binaries
```

Für `mdtoc` bedeutet das:

```text
mdtoc als native Go-CLI
→ gut für lokale Entwicklerumgebungen

mdtoc in github.dev
→ nur mit Einschränkungen oder anderer Architektur
```

Auch hier wäre für eine erste Extension-Version die sinnvolle Einschränkung:

```text
Support: VS Code Desktop
Kein Support: github.dev / vscode.dev
```

---

## Browser-Sandbox

Eine **Sandbox** ist eine kontrollierte, eingeschränkte Laufzeitumgebung.

Browser führen Code absichtlich stark eingeschränkt aus, damit Webseiten nicht einfach auf deinen Rechner zugreifen können.

Eine Browser-Sandbox verhindert zum Beispiel normalerweise:

```text
- beliebige lokale Programme starten
- direkt auf beliebige lokale Dateien zugreifen
- native Prozesse erzeugen
- Systembefehle ausführen
```

Das ist sicherheitstechnisch gewollt.

Für `mdtoc` ist das relevant, weil dein Tool aktuell ein natives Go-Programm ist. Eine Desktop-Extension kann dieses Programm starten; eine Browser-Extension nicht ohne Weiteres.

---

## CLI

**CLI** steht für **Command Line Interface**, also **Kommandozeilenprogramm**.

Dein `mdtoc` ist aktuell eine CLI. Man ruft es etwa so auf:

```bash
mdtoc generate README.md
mdtoc strip README.md
mdtoc regen README.md
mdtoc check README.md
```

Oder über Pipes:

```bash
cat README.md | mdtoc regen > README.new.md
```

Die CLI ist besonders wertvoll, weil sie unabhängig von VS Code funktioniert:

```text
- Terminal
- Shell-Skripte
- CI/CD
- Git Hooks
- Makefiles
- andere Editoren
```

Darum war die Empfehlung:

> CLI unbedingt behalten und die VS-Code-Extension nur als zusätzliche Bedienoberfläche darüber bauen.

---

## Extension

Eine **Extension** ist eine Erweiterung für VS Code.

Sie kann zum Beispiel:

```text
- neue Befehle hinzufügen
- Menüpunkte einfügen
- Tastenkürzel bereitstellen
- Dateien analysieren
- Warnungen anzeigen
- Formatierungen ausführen
- Einstellungen bereitstellen
```

Für `mdtoc` wäre die Extension nicht der eigentliche Markdown-Prozessor, sondern eher ein Adapter:

```text
VS-Code-Extension
→ liest aktives Markdown-Dokument
→ ruft mdtoc auf
→ übernimmt Ergebnis in den Editor
```

---

## Adapter

Ein **Adapter** ist eine dünne Vermittlungsschicht zwischen zwei Systemen.

In deinem Fall:

```text
VS Code API  ← Adapter →  mdtoc CLI
```

Die Extension müsste wissen:

```text
- welches Dokument ist gerade offen?
- ist es Markdown?
- welcher mdtoc-Befehl soll laufen?
- wo liegt das mdtoc-Binary?
- wie wird stdout/stderr verarbeitet?
- wie wird das Ergebnis zurück in den Editor geschrieben?
```

Sie müsste aber **nicht** selbst wissen, wie ein Inhaltsverzeichnis erzeugt wird. Das bleibt Aufgabe von `mdtoc`.

---

## Go-Binary

Ein **Binary** ist eine fertig kompilierte ausführbare Datei.

Bei Go ist das besonders praktisch, weil Go normalerweise einzelne ausführbare Dateien erzeugt:

```text
Linux:   mdtoc
macOS:   mdtoc
Windows: mdtoc.exe
```

Eine VS-Code-Extension könnte solche Binaries mitliefern, zum Beispiel:

```text
extension/
└─ bin/
   ├─ linux-x64/mdtoc
   ├─ linux-arm64/mdtoc
   ├─ darwin-x64/mdtoc
   ├─ darwin-arm64/mdtoc
   └─ win32-x64/mdtoc.exe
```

Dann müsste der User `mdtoc` nicht separat installieren.

---

## PATH

**PATH** ist eine Umgebungsvariable deines Betriebssystems.

Sie sagt dem System, wo es nach ausführbaren Programmen suchen soll.

Wenn du im Terminal eingibst:

```bash
mdtoc regen README.md
```

sucht das Betriebssystem in den Verzeichnissen aus `PATH` nach einem Programm namens `mdtoc`.

Für die Extension gibt es zwei Möglichkeiten:

```text
1. Sie sucht mdtoc im PATH.
2. Sie bringt ihr eigenes mdtoc-Binary mit.
```

Am robustesten wäre eine Kombination:

```text
1. User-spezifischer Pfad aus Extension-Setting
2. mitgeliefertes Binary
3. mdtoc aus PATH
4. Fehlermeldung mit Installationshinweis
```

---

## TypeScript

**TypeScript** ist eine Programmiersprache, die auf JavaScript aufbaut und Typen ergänzt.

VS-Code-Extensions werden typischerweise in TypeScript oder JavaScript geschrieben.

Das bedeutet aber nicht, dass dein Tool selbst nach TypeScript portiert werden muss.

Die Aufteilung wäre:

```text
Go:
- Markdown analysieren
- ToC erzeugen
- ToC entfernen
- ToC regenerieren
- Check ausführen

TypeScript:
- VS-Code-Befehl registrieren
- aktives Dokument lesen
- mdtoc-Prozess starten
- Ausgabe entgegennehmen
- Editor aktualisieren
```

---

## Node.js

**Node.js** ist eine JavaScript-Laufzeit außerhalb des Browsers.

VS-Code-Desktop-Extensions laufen in einer Umgebung, in der Node.js-Funktionalität verfügbar ist.

Das ist wichtig, weil Node.js Funktionen wie diese bietet:

```ts
child_process.execFile(...)
```

Damit kann die Extension ein externes Programm starten, also zum Beispiel dein Go-Binary `mdtoc`.

Im Browser gibt es diese Möglichkeit nicht.

---

## `child_process`

`child_process` ist ein Node.js-Modul zum Starten externer Prozesse.

Beispielhaft:

```ts
import { execFile } from "child_process";

execFile("mdtoc", ["regen"], ...);
```

Für `mdtoc` wäre das der zentrale Mechanismus, mit dem die VS-Code-Extension die bestehende Go-CLI verwendet.

---

## stdin, stdout, stderr

Diese drei Begriffe kommen aus der Unix-/CLI-Welt.

### stdin

**stdin** bedeutet **standard input**.

Das ist die Eingabe eines Programms.

Beispiel:

```bash
cat README.md | mdtoc regen
```

Hier bekommt `mdtoc` den Inhalt von `README.md` über stdin.

### stdout

**stdout** bedeutet **standard output**.

Das ist die normale Ausgabe eines Programms.

Beispiel:

```bash
mdtoc regen README.md > README.new.md
```

Das Ergebnis wird über stdout ausgegeben und in eine Datei geschrieben.

### stderr

**stderr** bedeutet **standard error**.

Dorthin gehören Fehlermeldungen, Warnungen und Diagnoseausgaben.

Für eine VS-Code-Extension ist eine saubere Trennung wichtig:

```text
stdout → nur der veränderte Markdown-Text
stderr → Fehlermeldungen, Warnungen, Diagnoseinformationen
```

Dann kann die Extension zuverlässig entscheiden, was sie in den Editor schreibt und was sie dem User als Fehler anzeigt.

---

## Exit Code

Ein **Exit Code** ist die Rückgabekennung eines Programms nach dem Beenden.

Typisch ist:

```text
0 = erfolgreich
nicht 0 = Fehler oder besonderer Zustand
```

Für `mdtoc check` wären differenzierte Exit Codes sinnvoll:

```text
0 = Dokument ist aktuell
1 = Dokument müsste aktualisiert werden
2 = echter Fehler, z. B. ungültige Eingabe
```

Das hilft der Extension, aber auch CI-Systemen.

---

## CI/CD

**CI/CD** steht für:

```text
CI = Continuous Integration
CD = Continuous Delivery / Continuous Deployment
```

Im Kontext von `mdtoc` heißt das:

```text
Bei jedem Commit oder Release automatisch:
- Tests ausführen
- Go-Binaries bauen
- VS-Code-Extension paketieren
- Release-Artefakte veröffentlichen
```

Zum Beispiel über GitHub Actions.

Für dein Projekt wäre CI/CD nützlich, um automatisch Binaries für mehrere Plattformen zu bauen:

```text
- Linux x64
- Linux arm64
- macOS Intel
- macOS Apple Silicon
- Windows x64
```

---

## VSIX

Eine **VSIX-Datei** ist das Paketformat für VS-Code-Extensions.

So ähnlich wie:

```text
.deb    für Debian/Ubuntu-Pakete
.msi    für Windows Installer
.vsix   für VS-Code-Extensions
```

Man kann eine VSIX lokal installieren oder im Marketplace veröffentlichen.

Für `mdtoc` könnte es später z. B. geben:

```text
mdtoc-0.1.0-linux-x64.vsix
mdtoc-0.1.0-darwin-arm64.vsix
mdtoc-0.1.0-win32-x64.vsix
```

---

## Marketplace

Der **Visual Studio Marketplace** ist die zentrale Plattform, über die VS-Code-Extensions veröffentlicht und installiert werden.

Wenn du deine Extension öffentlich verfügbar machen willst, wäre das der übliche Weg.

Für eine frühe Testphase brauchst du den Marketplace aber nicht zwingend. Du kannst auch lokal mit einer `.vsix` testen.

---

## `vsce`

**vsce** ist das Kommandozeilentool zum Paketieren und Veröffentlichen von VS-Code-Extensions.

Typische Befehle:

```bash
vsce package
vsce publish
```

Für platform-spezifische Pakete kann man Zielplattformen angeben, z. B. für Windows, Linux oder macOS.

---

## GoReleaser

**GoReleaser** ist ein Tool, das Go-Projekte automatisch für verschiedene Plattformen bauen und veröffentlichen kann.

Für `mdtoc` wäre GoReleaser interessant, wenn du automatisiert solche Artefakte erzeugen möchtest:

```text
- mdtoc für Linux
- mdtoc für macOS
- mdtoc.exe für Windows
- Checksums
- GitHub Release
```

Wenn du später VS-Code-Binaries mitliefern willst, kann GoReleaser Teil der Release-Pipeline sein.

---

## WebAssembly / WASM

**WebAssembly**, kurz **WASM**, ist ein Binärformat, mit dem Code aus Sprachen wie Go, Rust oder C im Browser laufen kann.

Für `mdtoc` wäre WASM relevant, wenn du die Logik in `vscode.dev` oder `github.dev` verwenden möchtest.

Dann wäre die Idee:

```text
Go-Code
→ nach WebAssembly kompilieren
→ im Browser ausführen
```

Das klingt attraktiv, ist aber aufwendiger als der CLI-Weg. Außerdem müssen Dateizugriff, stdin/stdout und Integration anders behandelt werden.

Für eine erste Version würde ich WASM nicht priorisieren.

---

## Diagnostics

**Diagnostics** sind die Warnungen, Fehler oder Hinweise, die VS Code im Editor anzeigen kann.

Beispiele:

```text
- rote Wellenlinie bei Fehlern
- gelbe Warnung bei veraltetem ToC
- Eintrag im "Problems"-Panel
```

Für `mdtoc check` könnte das später bedeuten:

```text
"Table of contents is outdated"
```

oder:

```text
"Heading level sequence may be inconsistent"
```

Im MVP wäre das nicht nötig. Später wäre es ein gutes Komfort-Feature.

---

## Code Action

Eine **Code Action** ist eine automatisch angebotene Aktion im Editor, oft über die Glühbirne.

Beispiel:

```text
Problem: ToC ist veraltet
Code Action: "Regenerate mdtoc table of contents"
```

Das wäre eine spätere, elegante Integration.

Für den Anfang reicht ein Command aus der Command Palette.

---

## Command Palette

Die **Command Palette** ist die zentrale Befehlsauswahl in VS Code.

Man öffnet sie typischerweise mit:

```text
Ctrl+Shift+P
```

oder auf macOS:

```text
Cmd+Shift+P
```

Dort könnte der User dann eingeben:

```text
mdtoc
```

und Befehle sehen wie:

```text
mdtoc: Generate Table of Contents
mdtoc: Regenerate Table of Contents
mdtoc: Strip Table of Contents
mdtoc: Check Table of Contents
```

---

## Workspace

Ein **Workspace** ist der aktuell geöffnete Projektkontext in VS Code.

Das kann sein:

```text
- ein einzelner Ordner
- mehrere Ordner
- ein Remote-Projekt
```

Wenn ich geschrieben habe „kein Workspace-weites Batch-Processing im MVP“, meinte ich:

> Die erste Version sollte nicht sofort alle Markdown-Dateien im ganzen Projekt bearbeiten.

Stattdessen:

```text
Nur die aktuell offene Markdown-Datei bearbeiten.
```

Das ist sicherer und einfacher.

---

## Format on Save

**Format on Save** bedeutet:

> Beim Speichern einer Datei wird automatisch eine Formatierung oder Transformation ausgeführt.

Für `mdtoc` könnte das heißen:

```text
User speichert README.md
→ Extension ruft automatisch mdtoc regen aus
→ ToC wird aktualisiert
```

Das ist praktisch, aber auch riskant:

```text
- unerwartete Änderungen
- Konflikte mit anderen Formatierern
- Performance bei großen Dateien
- schwerer nachvollziehbare Git-Diffs
```

Darum sollte es nicht standardmäßig aktiv sein.

Empfehlung:

```text
mdtoc.formatOnSave = false
```

und User können es bewusst aktivieren.

---

## Virtual Workspace

Ein **Virtual Workspace** ist ein Arbeitsbereich, bei dem Dateien nicht normal auf dem lokalen Dateisystem liegen.

Beispiele können sein:

```text
- Browser-basierte Repositories
- Remote-Dateisysteme
- schreibgeschützte Quellen
- virtuelle Dokumente von Extensions
```

Für ein Tool, das ein lokales Binary startet und Dateien verarbeitet, sind Virtual Workspaces komplizierter.

Darum: im MVP nicht unterstützen oder nur sehr eingeschränkt.

---

## Native Extension vs. Web Extension

Eine **Native/Desktop Extension** läuft in VS Code Desktop und kann Node.js-Funktionen verwenden.

Eine **Web Extension** läuft im Browser und ist deutlich eingeschränkter.

Für `mdtoc`:

```text
Desktop Extension:
- kann Go-Binary starten
- guter Fit für bestehende CLI
- einfachster Weg

Web Extension:
- kann kein Go-Binary starten
- braucht TypeScript-Core oder WASM
- deutlich mehr Aufwand
```

---

## Zusammenfassung für `mdtoc`

Die wichtigsten Begriffe in deiner Entscheidung sind diese:

```text
MVP
= kleine erste Version mit Generate/Regen/Strip/Check für aktive Markdown-Datei

CLI
= bestehendes Go-Kommandozeilentool, bleibt erhalten

VS-Code-Extension
= TypeScript-Adapter, der die CLI aus VS Code heraus bedienbar macht

Go-Binary
= kompilierte mdtoc-Datei, die die Extension starten kann

vscode.dev / github.dev
= Browser-Varianten, in denen native Go-Binaries nicht direkt laufen

WASM
= mögliche spätere Lösung für Browser-Support

VSIX
= installierbares Paket der Extension

Marketplace
= öffentliche Veröffentlichungsplattform für VS-Code-Extensions
```

## Technische Empfehlung

```text
Erste Version:
VS Code Desktop Extension + bestehende Go-CLI

Später optional:
WASM oder TypeScript-Port für vscode.dev/github.dev
```
