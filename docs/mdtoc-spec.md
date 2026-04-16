# mdtoc – Spezifikation (v1)

## 1. Zweck und Grundprinzipien

`mdtoc` ist ein deterministisches CLI-Werkzeug zur Verarbeitung einzelner Markdown-Dokumente.

Funktionen:
- Generierung eines Inhaltsverzeichnisses (ToC)
- konsistente Kapitelnummerierung
- Erzeugung stabiler Anchor-IDs, unabhängig von Kapitelnummern
- Entfernen aller von `mdtoc` generierten Artefakte
- Zustandsprüfung eines Dokuments für CI

Grundprinzipien:
- Der sichtbare Überschriftentext ist die einzige semantische Quelle der Wahrheit.
- Kapitelnummern sind abgeleitet und nicht persistent.
- Anchor-IDs werden nur aus dem unnummerierten Titel berechnet.
- Generierte Inhalte sind vollständig rekonstruierbar.
- `mdtoc` ändert ein Dokument nur auf Basis einer klar definierten verwalteten Struktur.
- Das Werkzeug ist idempotent.

_Anmerkung:_ "formal" bedeutet in diesem Dokument nur "klar genug für Parser, Tests und spätere Code-Generierung". Gemeint ist keine große Architektur, sondern ein kleiner, robuster Vertragsrahmen.

## 2. Geltungsbereich und Nicht-Ziele

`mdtoc` verarbeitet absichtlich nur einen kleinen, eindeutigen Markdown-Subset.

Unterstützt in v1:
- einzelne Markdown-Datei
- ATX-Überschriften mit `#` bis `#########`
- definierte ToC-Marker
- definierter Config-Block
- definierte Inline-Ankerform in Überschriften

Nicht unterstützt in v1:
- Setext-Überschriften
- GUI-Automation
- PDF-Erzeugung
- Mehrdateiverarbeitung
- vollständiger Markdown-AST
- partielle Verarbeitung wie `--toc-only` oder `--anchors-only`

_Anmerkung:_ Die Beschränkung auf einen kleinen Markdown-Subset ist Absicht. Damit bleiben Parser, Testfälle und Fehlersuche einfach.

## 3. Explizite Dokumentstruktur

Ein von `mdtoc` verwaltetes Dokument verwendet genau diese Container-Struktur:

```md
<!-- mdtoc -->
[TOC CONTENT]
<!-- mdtoc-config
numbering=on
min-level=2
max-level=4
anchors=on
toc=on
state=generated
-->
<!-- /mdtoc -->
```

Regeln:
- Der äußere Container besteht aus Start-Marker, ToC-Bereich, Config-Block und End-Marker.
- Der Config-Block muss unmittelbar vor `<!-- /mdtoc -->` stehen.
- `<!-- mdtoc -->` darf höchstens einmal vorkommen.
- `<!-- /mdtoc -->` darf höchstens einmal vorkommen.
- Der Config-Block darf höchstens einmal vorkommen.
- Kommt keiner der äußeren Marker vor, fügt `generate` den kompletten Container am Dateianfang ein.
- Kommt nur einer der äußeren Marker vor oder liegt der Start-Marker nach dem End-Marker, ist das ein Parsing-Fehler.
- Alles zwischen `<!-- mdtoc -->` und dem Anfang des Config-Blocks ist der verwaltete ToC-Bereich.
- Fremder Inhalt im ToC-Bereich wird bei `generate` nicht gelöscht, sondern als HTML-Kommentar erhalten.

Hinweis: Der User kann durch Verschieben des Toc-Bereiches bestimmen, wo das Inhaltsverzeichnis sein soll.

_Anmerkung:_ Die explizite Container-Struktur ist absichtlich einfacher lesbar als eine implizite Marker-Logik. So ist sofort sichtbar, welchen Bereich `mdtoc` verwaltet.

## 4. Parsing-Regeln

### 4.1 Grundsatz

Der Parser arbeitet zeilenbasiert.  
Er erkennt nur die in dieser Spezifikation genannten Strukturen und ignoriert bewusst andere Markdown-Sonderfälle.

### 4.2 Ignored Regions

Diese Bereiche werden beim Erkennen von Markern und Überschriften ignoriert:

1. Fenced code blocks mit Backticks:
   - Beginn: eine Zeile, die mit drei Backticks beginnt
   - Ende: die nächste Zeile, die mit drei Backticks beginnt
2. Fenced code blocks mit Tilde:
   - Beginn: eine Zeile, die mit drei Tilden (`~~~`) beginnt
   - Ende: die nächste Zeile, die mit drei Tilden (`~~~`) beginnt
3. Inline code spans:
   - Bereich zwischen zwei Backticks in derselben Zeile
4. HTML-Kommentare:
   - `<!-- ... -->`
   - Ausnahme: `<!-- mdtoc -->`, `<!-- /mdtoc -->` und `<!-- mdtoc-config ... -->`

Nicht ignoriert:
5. Blockquotes

Blockquotes sind normale Eingabezeilen.  
Sie werden nicht als Sonderbereich behandelt.

Praktische Folge:
- Eine Blockquote-Zeile beginnt mit optionalen Spaces und dann `>`.
- Eine von `mdtoc` erkannte Überschrift muss mit einem `hashes`-Präfix direkt an Spalte 1 beginnen.
- Dadurch können Blockquotes nicht auf die Heading-Syntax matchen und brauchen keine Sonderbehandlung.

### 4.3 Parsing-Reihenfolge

Die Verarbeitung läuft in dieser Reihenfolge:

1. Äußeren `mdtoc`-Container und Config-Block erkennen.
2. Ignored Regions berücksichtigen.
3. Überschriften erkennen.
4. Verwaltete Artefakte semantisch normalisieren.
5. Sollzustand ableiten.
6. Ausgabe rendern.

## 5. Heading-Syntax

### 5.1 Kandidaten für Überschriften

Nur Zeilen, die direkt am Zeilenanfang mit einem der folgenden Präfixe beginnen, sind für `mdtoc` überhaupt Überschriften:

```text
hashes := "# " | "## " | "### " | "#### " | "##### " | "###### "| "####### "| "######## "| "######### "
```

Damit gilt zugleich:
- nach den `#` muss genau ein Leerzeichen folgen
- vor den `#` dürfen keine Leerzeichen stehen

_Anmerkung:_ Das Leerzeichen ist hier bewusst Teil von `hashes`.  Das vereinfacht den Parser: Nach dem Präfix kommt entweder direkt die Nummer, direkt der Anchor oder direkt der Titel.

### 5.2 Struktur einer verwalteten Überschrift

Verwaltete Überschriften verwenden genau dieses Schema:

```text
heading_line := hashes [number SP] [anchor] title
number       := DIGIT+ ("." DIGIT+)* "."
anchor       := "<a id=\"anchor_id\"></a>"
title        := NONEMPTY_TEXT
SP           := exactly one U+0020 space
```

Zusätzliche Regeln:
- `number` ist optional.
- Wenn `number` vorkommt, steht sie direkt nach `hashes` und wird von genau einem Leerzeichen gefolgt.
- `anchor` ist optional.
- Wenn `anchor` vorkommt, steht er direkt nach `hashes` oder direkt nach `number SP`.
- Zwischen `</a>` und dem ersten Zeichen des Titels steht **kein** Leerzeichen.
- Diese fehlende Leerstelle ist absichtlich so festgelegt, um mit `dumeng-toc` kompatibel zu bleiben.
- Innerhalb des Titels bleiben Leerzeichen und Zeichen unverändert erhalten.
- Nur Überschriften, die exakt dieser Positionslogik entsprechen, dürfen von `mdtoc` umgeschrieben werden.

Beispiele gültiger verwalteter Überschriften:

```md
# Titel
## 1. Einführung
## <a id="einfuehrung"></a>Einführung
### 2.1. <a id="api-overview"></a>API Overview
```

Beispiele, die `mdtoc` nicht als verwaltete Struktur behandelt:

```md
 # Titel
##  1. Einführung
### 1.2 Einführung
### <a id="x"></a> Einführung
```

### 5.3 Bedeutung der Syntax

- `### 2024 roadmap` ist **keine** Nummer, weil das erste Token nicht auf `.` endet.
- `### 3D graphics` ist **keine** Nummer, weil das erste Token kein reines `x.y.z.`-Muster ist.
- `### 2.1. API` ist eine verwaltete Nummernsyntax.

_Anmerkung:_ Das Muster `### 2.1. API` ist damit bewusst für `mdtoc` reserviert. Wer eine freie Überschrift exakt in diesem Format schreibt, verwendet dieselbe Syntax wie das Tool.

### 5.4 Unterstützter Markdown-Subset

`mdtoc` ist kein allgemeiner Markdown-Parser.

Für Überschriften gilt in v1:
- nur ATX-Überschriften
- nur die oben definierte Heading-Syntax
- keine Setext-Überschriften
- keine impliziten oder mehrdeutigen Sonderfälle

Der praktische Vorfilter lautet damit mindestens:

```text
^#{1,9} 
```

Und die eigentliche Umschreibelogik greift nur auf Zeilen, die auch die restliche Positionslogik erfüllen.

## 6. Kleines formales Modell

Dieser Abschnitt beschreibt die minimale interne Sicht, die für saubere Implementierung und Tests hilfreich ist.

### 6.1 Verwaltete Überschrift

Intern reicht für eine verwaltete Überschrift dieses Modell:

```text
ManagedHeading
- line_index
- level
- title
- number        // abgeleitet oder leer
- anchor_id     // abgeleitet oder leer
```

Semantisch wichtig sind nur:
- `level`
- `title`

Abgeleitet werden daraus:
- `number`
- `anchor_id`

### 6.2 Dokumentzustand

Ein Dokument befindet sich für `mdtoc` praktisch in einem dieser Zustände:

- `unmanaged`  
  Kein gültiger `mdtoc`-Container mit gültigem Config-Block vorhanden.

- `generated`  
  Das Dokument entspricht dem aus Inhalt + Config abgeleiteten Sollzustand.

- `stripped`  
  Die verwalteten Artefakte sind entfernt, der Container mit Config bleibt erhalten.

### 6.3 Verarbeitungspipeline

Die Verarbeitung folgt immer demselben einfachen Muster:

```text
parse -> normalize -> derive -> render
```

Das bedeutet:
- **parse**: Container, Config und Überschriften erkennen
- **normalize**: verwaltete Nummern und verwaltete Anchors semantisch entfernen
- **derive**: Nummern, Anchor-IDs und ToC neu berechnen
- **render**: Dokument deterministisch zurückschreiben

_Anmerkung:_ Das soll keine große AST-Architektur erzwingen. Es legt nur fest, welche Informationen semantisch zählen und welche nur Render-Artefakte sind.

### 6.4 Gültigkeitsbereich von `min-level` und `max-level`

Diese Fassung geht von folgender, leicht verständlicher Regel aus:

- `min-level` und `max-level` filtern dieselbe Menge an Überschriften für
  - ToC-Erzeugung
  - Nummerierung
  - Anchor-Erzeugung

Praktische Folge:
- Bei `generate` werden zunächst alle verwalteten Nummern und verwalteten Anchors aus allen verwalteten Überschriften entfernt.
- Danach werden Nummern und Anchors nur für Überschriften innerhalb des aktiven Level-Bereichs neu gesetzt.
- Überschriften außerhalb des Bereichs bleiben inhaltlich erhalten, werden aber nicht neu verwaltet.

## 7. Config-Block

Der Config-Block hat genau diese Form:

```html
<!-- mdtoc-config
numbering=on
min-level=2
max-level=4
anchors=on
toc=on
state=generated
-->
```

Regeln:
- Der Config-Block ist zeilenbasiert.
- Jedes Feld steht in einer eigenen Zeile.
- Alle Felder verwenden `key=value`.
- Die Feldreihenfolge ist fest:
  1. `numbering`
  2. `min-level`
  3. `max-level`
  4. `anchors`
  5. `state`
- Zulässige Werte:
  - `numbering=on|off`
  - `anchors=on|off`
  - `state=generated|stripped`
- `min-level` und `max-level` sind positive ganze Zahlen.
- `min-level` darf nicht größer als `max-level` sein.
- `max-level` darf nicht größer als 9 sein.
- `generate` schreibt alle Generator-Optionen in den Config-Block; nicht angegebene Optionen mit Default-Wert.
- `--file`, `--help`, `--version`, `--verbose` und `--raw` werden nicht persistiert.
- `strip` behält den Config-Block und setzt nur `state=stripped`.
- `strip --raw` entfernt den Config-Block vollständig.

_Anmerkung:_ `state` ist hier bewusst ebenfalls auf `key=value` vereinheitlicht. Das macht den Parser trivialer und vermeidet einen unnötigen Sonderfall.

## 8. CLI-Schnittstelle

### 8.1 Kommandos

| Option                                  | Beschreibung                                     |
|-----------------------------------------|--------------------------------------------------|
| `mdtoc --version`                       | Gibt kurze Versionsinfo aus.                     |
| `mdtoc --version --verbose`             | Gibt ausführliche Versionsinfo aus.              |
|                                         |                                                  |
| `mdtoc --help`                          | Gibt kurzen Help Text aus aus.                   |
| `mdtoc --help --verbose`                | Gibt langen Help Text aus aus.                   |
|                                         |                                                  |
| `mdtoc generate [--verbose] [OPTIONEN]` | generiert/updated ToC, numbers, anchors.         |
| `mdtoc generate  --help`                | Gibt langen Help Text speziell für generate aus. |
|                                         |                                                  |
| `mdtoc strip    [--verbose] [--raw]`    | entfernt ToC, numbers, anchors und ggf. Config.  |
| `mdtoc strip     --help`                | Gibt langen Help Text speziell für strip aus.    |
|                                         |                                                  |
| `mdtoc check    [--verbose]`            | prüft Config und ggf. ToC, numbers, anchors.     |
| `mdtoc check     --help`                | Gibt langen Help Text speziell für check aus.    |

### 8.2 Optionen für `generate`

| Option                  | Default | Bedeutung                                                |
|-------------------------|---------|----------------------------------------------------------|
| `--numbering <on\|off>` | `on`    | Kapitelnummern aktivieren oder deaktivieren              |
| `--min-level <N>`       | `2`     | minimale verwaltete Heading-Ebene (>=1)                  |
| `--max-level <N>`       | `4`     | maximale verwaltete Heading-Ebene (<=9)                  |
| `--anchors <on\|off>`   | `on`    | Inline-Anker erzeugen oder deaktivieren                  |
| `--toc <on\|off>`       | `on`    | Schreibt [TOC CONTENT], wenn `on` oder nicht, wenn `off` |
| `--file <name>`         | –       | Datei lesen und überschreiben                            |
| `--verbose`             | `off`   | Diagnose- und Ablaufmeldungen auf `stderr`               |
| `--help`                | –       | Hilfe anzeigen                                           |

Kurzformen:

| Option        | Kurzform |
|---------------|----------|
| `--numbering` | `-n`     |
| `--anchors`   | `-a`     |
| `--file`      | `-f`     |
| `--verbose`   | `-v`     |
| `--help`      | `-h`     |

### 8.3 I/O- und Logging-Verhalten

- Mit `--file` wird die Datei gelesen und überschrieben.
- Ohne `--file` kommt die Eingabe von `stdin` und die Dokumentausgabe geht auf `stdout`.
- Erfolgreiche Kommandos erzeugen keine Ausgabe, außer bei `--help`, `--version` oder `--verbose`.
- Fehler und Diagnosemeldungen gehen ausschließlich auf `stderr`.

## 9. Kommandos

### 9.1 `generate`

Verhalten:
1. Dokument parsen.
2. Wenn kein verwalteter Container vorhanden ist, kompletten Container am Dateianfang anlegen.
3. Wenn Marker-Struktur oder Config ungültig ist: Fehler und keine Änderung.
4. Vorhandene verwaltete Artefakte semantisch entfernen:
   - ToC-Inhalt
   - verwaltete Kapitelnummern
   - verwaltete Inline-Anker
5. Relevante Überschriften bestimmen.
6. Nummern neu berechnen, falls `numbering=on`.
7. Anchor-IDs neu berechnen, falls `anchors=on`.
8. ToC neu rendern.
9. Überschriften neu rendern.
10. Config neu rendern und `state=generated` setzen.
11. Dokument zurückschreiben.

Zusätzliche Regeln:
- Nummerierung und Anchor-ID sind strikt entkoppelt.
- Anchor-IDs werden nur aus dem unnummerierten Titel berechnet.
- Duplicate IDs werden deterministisch aufgelöst.
- Fremder Inhalt im ToC-Bereich wird nicht gelöscht, sondern als HTML-Kommentar erhalten.
- Bei Erfolg ist das Ergebnis idempotent.

Beispiel für eine gerenderte Überschrift:

```md
### 4.1. <a id="open-source"></a>Open  source <a id="opensource"></a>
```
_Anmerkung:_ Die gestrippte Überschrift ist hier `### Open  source <a id="opensource"></a>`, kann also mehrere Leerzeichen und auch weitere Links enthalten, wobei diese **nicht vor dem sichtbaren Titel** stehen dürfen um mdtoc einfacher idempotent zu machen.  

### 9.2 `strip`

Verhalten:
- benötigt einen gültigen Config-Block
- entfernt verwalteten ToC-Inhalt
- entfernt verwaltete Kapitelnummern
- entfernt verwaltete Inline-Anker
- behält äußeren Container
- behält Config-Block
- setzt `state=stripped`

Nach `strip` ist damit diese Struktur weiterhin gültig:

```md
<!-- mdtoc -->
<!-- mdtoc-config
numbering=on
min-level=2
max-level=4
anchors=on
toc=on
state=stripped
-->
<!-- /mdtoc -->
```

Fehlerfall:
- kein gültiger Config-Block -> Fehler
- keine implizite Reparatur

### 9.3 `strip --raw`

Verhalten:
- ignoriert den Config-Block
- entfernt den kompletten verwalteten Container, falls vorhanden:
  - `<!-- mdtoc -->`
  - ToC-Inhalt
  - `mdtoc-config`
  - `<!-- /mdtoc -->`
- entfernt zusätzlich verwaltete Kapitelnummern
- entfernt zusätzlich verwaltete Inline-Anker

Konservative Regel:
- Wenn nicht sicher entschieden werden kann, ob eine Nummer oder ein Inline-Anker verwaltet ist, bleibt der Inhalt stehen.
- Dieser Fall soll in jedem Fall als Diagnose ausgegeben werden.

Einsatzfälle:
- beschädigte Config
- Migration
- vollständige Entfernung der `mdtoc`-Verwaltung
- Tests

### 9.4 `check`

Verhalten:
- benötigt einen gültigen Config-Block
- rekonstruiert den Sollzustand aus aktuellem Dokumentinhalt und Config
- vergleicht Sollzustand und Istzustand byte-genau
- liefert `0`, wenn beide identisch sind
- liefert bei Abweichung einen Fehler-Exit-Code

Keine Seiteneffekte:
- `check` verändert das Dokument nie

_Anmerkung:_ "byte-genau" klingt formaler als es praktisch ist. Gemeint ist: `check` berechnet denselben Text, den `generate` oder `strip` schreiben würden, und vergleicht genau diesen.

## 10. ToC-Regeln

Der ToC basiert auf allen verwalteten Überschriften innerhalb von `min-level` bis `max-level` inklusive.

Render-Regeln:
- Eine Überschrift erzeugt genau einen ToC-Eintrag.
- Die Hierarchie folgt dem Heading-Level.
- Pro zusätzlicher Ebene relativ zu `min-level` wird um zwei Spaces eingerückt.
- Jeder Eintrag ist ein Markdown-Listenpunkt mit Link.

Beispiel:

```md
* [1. Einleitung](#einleitung)
  * [1.1. API](#api)
```

Anzeige im Linktext:
- bei `numbering=on`: `nummer + titel`
- bei `numbering=off`: nur `titel`

Linkziel:
- grundsätzlich die von `mdtoc` berechnete Anchor-ID

Was genau bedeutet `anchors=off` für die ToC-Links?
- Wenn `anchors=off` keine Inline-Anker schreibt, hängt die Zielauflösung sonst vom Markdown-Renderer ab.  
- Für wirklich stabile Links muss `anchors=on` verwendet werden.
- Ansonsten ist das Funktionieren der Links nicht garantiert in allen Markdown-Renderent.

## 11. Anchor-Generierung (normativ)

_Designentscheidung:_ mdtoc verwendet standardmäßig slug-basierte Anchor-IDs, da diese stabil, URL-sicher und mit gängigen Markdown-Renderern kompatibel sind.

Anchor-IDs werden deterministisch aus dem **unnummerierten Überschriften-Text (`title`)** erzeugt.

### 11.1 Ziel

Die erzeugten Anchor-IDs sollen sein:

- stabil
- deterministisch
- URL-tauglich
- CI-freundlich
- plattformunabhängig

### 11.2 Algorithmus

Für jede Überschrift gilt:

```text
anchor_id := slugify(title)
```

### 11.3 Slug-Regeln (normativ)

Die Funktion `slugify` MUST wie folgt arbeiten:

1. Eingabe ist der unveränderte Überschriften-Text (`title`)
2. Konvertiere den Text zu lowercase
3. Ersetze deutsche Umlaute und `ß` wie folgt:
   - `ä -> ae`
   - `ö -> oe`
   - `ü -> ue`
   - `Ä -> ae`
   - `Ö -> oe`
   - `Ü -> ue`
   - `ß -> ss`
4. Zerlege Unicode-Zeichen mit Diakritika in Grundzeichen + kombinierende Zeichen
5. Entferne kombinierende Zeichen  
   Beispiel:
   - `é -> e`
   - `ñ -> n`
   - `ç -> c`
6. Ersetze jede Folge aus einem oder mehreren Zeichen, die **nicht** `[a-z0-9]` sind, durch genau ein `-`
7. Entferne führende und folgende `-`
8. Falls das Ergebnis leer ist, ist dies ein Fehler
9. Kollisionen werden deterministisch mit Suffixen `-2`, `-3`, ... aufgelöst

### 11.4 Beispiele

#### Beispiel 1

```md
### Open source
```

→

```text
open-source
```

#### Beispiel 2

```md
### Übergrößenträger & naïve façade – déjà vu!
```

→

```text
uebergroessentraeger-naive-facade-deja-vu
```

#### Beispiel 3

```md
### Ä Ö Ü ä ö ü ß
```

→

```text
ae-oe-ue-ae-oe-ue-ss
```

#### Beispiel 4

```md
### 中文 русский عربى
```

→ Fehler, falls nach Normalisierung kein `[a-z0-9]` übrig bleibt


### 11.5 Kollisionen

Kollisionen werden in Auftretensreihenfolge aufgelöst.

Beispiel:

- `overview`
- `overview-2`
- `overview-3`

Die Kollisionsauflösung MUST auf dem finalen Slug erfolgen.

### 11.6 Begründung

Diese Regeln sind bewusst:

- einfach zu implementieren
- leicht testbar
- stabil über Betriebssysteme hinweg
- für deutschsprachige Dokumente gut lesbar

### 11.7 Designentscheidung

`mdtoc` verwendet für Version 1 standardmäßig diese Slug-Strategie.

```text
default anchor-style = slug
```

_Anmerkung:_ Das ist für Go angenehm, weil es in klaren Schritten gebaut werden kann.

- lowercase
- deutsche Sonderfälle ersetzen
- Unicode-Normalisierung
- Regex für Nicht-Alphanumerisches
- trim
- Kollisionen

### 11.8 Sonderfall

- Ein komplett nicht-lateinischer Titel soll kein Fehler sein. Dann einen Fallback wie section, section-2, ... erzeugen.

_Anmerkung:_ Es gibt viele User in China.

## 12. Fehlerverhalten, Logging und Exit-Codes

Fehlerfälle:
- fehlender oder unvollständiger `mdtoc`-Container
- fehlender Config-Block bei `strip` oder `check`
- ungültiger Config-Block
- Parsing-Fehler
- ungültige Optionen

Grundregeln:
- Fehler werden auf `stderr` ausgegeben.
- Bei Fehlern gibt es keine implizite Reparatur, außer dem explizit erlaubten Anlegen eines neuen Containers durch `generate`, wenn noch gar keine `mdtoc`-Verwaltung existiert.
- Erfolgreiche Kommandos schreiben keine Statusmeldungen auf `stdout`.

Empfohlene Exit-Codes:
- `0` -> Erfolg
- `1` -> Parsing-, Config- oder CLI-Fehler
- `2` -> `check` hat eine Abweichung gefunden

## 13. Idempotenz

Idempotenz ist Teil des Vertrags.

Beispiele:

```bash
mdtoc generate
mdtoc generate
```

=> keine weitere Änderung beim zweiten Lauf

```bash
mdtoc strip
mdtoc strip
```

=> keine weitere Änderung beim zweiten Lauf

```bash
mdtoc strip --raw
mdtoc strip --raw
```

=> keine weitere Änderung beim zweiten Lauf

## 14. Erweiterbarkeit

Mögliche spätere Erweiterungen:
- alternative Anchor-Styles
- alternative ToC-Formate
- Versionierung im Config-Block
- weitere Ausgabeformate

_Anmerkung:_ Diese Punkte sind ausdrücklich Erweiterungen. Sie sollen v1 nicht unnötig komplex machen.

---

- Analysiere mdtoc-spec.md auf Korrektheit, Vollständigkeit, Inkonsistenzen, Widersprüche, Unklarheiten (Interpretationsspielraum).
- Nimm notwendige Änderungen vor und kommentiere zu entfernenden Text aus, statt zu löschen, damit das Review leichter ist.
- Versuche nicht, den Text einfach nur anders zu schreiben. Wenn wirklicher Mehrwert möglich ist, dann ja.
- Schreibe Deine Interpretationen (falls nötig), Erklärungen und Rückfragen als Blockquotes in die Datei.
- Ziel ist es den Text als Implementations-Spezifikation zur Programm-Generierung zu verwenden.
- Wenn kleinere Widerholungen (Inhalts-Doppelungen) auftreten, prüfe ob widerspruchsfrei und verweise jeweils auf die andere Stelle.
- Da die Zielsprache Go ist, prüfe ob die Verwendung das goldmark Packages sinnvoll ist (mdtoc Code-Reduzierung) und welche Version/Variante in Frage kommt.
- Passe die interne slug-Spezifikation und die interne Anchor-Spezifikation so an, dass sie quasi gleich und Github-Kompatibel sind. Idealerweise eine einzige Spezifikation für die Gemeinsamkeiten

