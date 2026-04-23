# mdtoc – Spezifikation (v1)

## 1. Zweck und Grundprinzipien

`mdtoc` ist ein deterministisches CLI-Werkzeug zur Verarbeitung einzelner Markdown-Dokumente.

Funktionen:

* Generierung eines Inhaltsverzeichnisses (ToC)
* konsistente Kapitelnummerierung
* Erzeugung stabiler Anchor-IDs, unabhängig von Kapitelnummern
* Entfernen aller von `mdtoc` generierten Artefakte
* Zustandsprüfung eines Dokuments für CI

Grundprinzipien:

* Der sichtbare Überschriftentext ist die einzige semantische Quelle der Wahrheit.
* Kapitelnummern sind abgeleitet und nicht persistent.
* Anchor-IDs werden nur aus dem unnummerierten Titel berechnet.
* Generierte Inhalte sind vollständig rekonstruierbar.
* `mdtoc` ändert ein Dokument nur auf Basis einer klar definierten verwalteten Struktur.
* Das Werkzeug ist idempotent.

_Anmerkung:_ "formal" bedeutet in diesem Dokument nur "klar genug für Parser, Tests und spätere Code-Generierung". Gemeint ist keine große Architektur, sondern ein kleiner, robuster Vertragsrahmen.

## 2. Geltungsbereich und Nicht-Ziele

`mdtoc` verarbeitet absichtlich nur einen kleinen, eindeutigen Markdown-Subset.

Unterstützt in v1:

* einzelne Markdown-Datei
* ATX-Überschriften mit `#` bis `######`
* definierte ToC-Marker
* definierter Config-Block
* definierte Inline-Ankerform in Überschriften

Nicht unterstützt in v1:

* Setext-Überschriften
* GUI-Automation
* PDF-Erzeugung
* Mehrdateiverarbeitung
* vollständiger Markdown-AST als eigener `mdtoc`-Spezifikationsgegenstand
* partielle Verarbeitung wie `--toc-only` oder `--anchors-only`

_Anmerkung:_ Die Beschränkung auf einen kleinen Markdown-Subset ist Absicht. Damit bleiben Parser, Testfälle und Fehlersuche einfach.

## 3. Explizite Dokumentstruktur

Ein von `mdtoc` verwaltetes Dokument verwendet genau diese Container-Struktur:

```md
<!-- mdtoc -->
[TOC CONTENT]
<!-- mdtoc-config
container-version=v2
numbering=true
min-level=2
max-level=4
anchor=github
toc=true
state=generated
-->
<!-- /mdtoc -->
```

Regeln:

* Der äußere Container besteht aus Start-Marker, ToC-Bereich, Config-Block und End-Marker.
* Der Config-Block muss unmittelbar vor `<!-- /mdtoc -->` stehen.
* `<!-- mdtoc -->` darf höchstens einmal vorkommen.
* `<!-- /mdtoc -->` darf höchstens einmal vorkommen.
* Der Config-Block darf höchstens einmal vorkommen.
* Kommt keiner der äußeren Marker vor, fügt `generate` den kompletten Container am Dateianfang ein.
* Kommt nur einer der äußeren Marker vor oder liegt der Start-Marker nach dem End-Marker, ist das ein Parsing-Fehler.
* Alles zwischen `<!-- mdtoc -->` und dem Anfang des Config-Blocks ist der verwaltete ToC-Bereich.
* Fremder Inhalt im ToC-Bereich wird bei `generate` nicht gelöscht, sondern als HTML-Kommentar erhalten.

Hinweis: Der User kann durch Verschieben des Toc-Bereiches bestimmen, wo das Inhaltsverzeichnis sein soll.

_Erklärung:_ Der komplette Container ist der verwaltete Bereich. `toc=off` bedeutet nicht "kein Container", sondern "ein leerer verwalteter ToC-Bereich".

_Anmerkung:_ Die explizite Container-Struktur ist absichtlich einfacher lesbar als eine implizite Marker-Logik. So ist sofort sichtbar, welchen Bereich `mdtoc` verwaltet.

## 4. Parsing-Regeln

### 4.1 Grundsatz

Die Spezifikation beschreibt das verwaltete Verhalten zeilen- und positionsbezogen.  
Eine Implementierung DARF intern einen Markdown-Parser verwenden, solange das externe Verhalten exakt dieser Spezifikation entspricht.

_Erklärung:_ Für die Implementierung in Go ist ein interner Parser wie `goldmark` sinnvoll, obwohl die verwalteten Umschreiberegeln weiterhin zeilenorientiert beschrieben bleiben.

### 4.2 Ignored Regions

Diese Bereiche werden beim Erkennen von Markern und Überschriften ignoriert:

1. Fenced code blocks mit Backticks:
   * Beginn: ein Backtick-Fence gemäß unterstütztem Markdown-Parser oder unterstütztem v1-Subset (eine Zeile, die mit drei Backticks beginnt)
   * Ende: das zugehörige schließende Backtick-Fence (die nächste Zeile, die mit drei Backticks beginnt)
2. Fenced code blocks mit Tilde:
   * Beginn: ein Tilde-Fence gemäß unterstütztem Markdown-Parser oder unterstütztem v1-Subset (eine Zeile, die mit drei Tilden (`~~~`) beginnt)
   * Ende: das zugehörige schließende Tilde-Fence (die nächste Zeile, die mit drei Tilden (`~~~`) beginnt)
3. Inline code spans:
   * Bereich zwischen zwei Backticks in derselben Zeile
4. HTML-Kommentare:
   * `<!-- ... -->`
   * Ausnahme: `<!-- mdtoc -->`, `<!-- /mdtoc -->` und `<!-- mdtoc-config ... -->`

Nicht ignoriert:

5. Blockquotes

Blockquotes sind normale Eingabezeilen.  
Sie werden nicht als Sonderbereich behandelt.

Praktische Folge:

* Eine Blockquote-Zeile beginnt mit optionalen Spaces und dann `>`.
* Eine von `mdtoc` erkannte Überschrift muss mit einem `hashes`-Präfix direkt an Spalte 1 beginnen.
* Dadurch können Blockquotes nicht auf die Heading-Syntax matchen und brauchen keine Sonderbehandlung.

_Interpretation:_

* "Blockquotes nicht ignorieren" bedeutet hier ausdrücklich nicht, dass daraus Überschriften entstehen.
* Es bedeutet nur, dass `mdtoc` keinen eigenen Blockquote-Modus braucht.

### 4.3 Parsing-Reihenfolge

Die Verarbeitung läuft logisch in dieser Reihenfolge:

1. Ignored Regions bzw. Markdown-Kontext bestimmen.
2. Äußeren `mdtoc`-Container und Config-Block nur außerhalb ignorierter Bereiche erkennen.
3. Überschriften nur außerhalb ignorierter Bereiche erkennen.
4. Verwaltete Artefakte semantisch normalisieren.
5. Sollzustand ableiten.
6. Ausgabe rendern.

_Erklärung:_

* Ohne diese Reihenfolge wären Marker oder Überschriften innerhalb eines Code-Fence mehrdeutig.
* Genau diese Unklarheit soll hier ausgeschlossen werden.

## 5. Heading-Syntax

### 5.1 Kandidaten für Überschriften

Nur Zeilen, die direkt am Zeilenanfang mit einem der folgenden Präfixe beginnen, sind für `mdtoc` überhaupt Überschriften:

```text
hashes := "# " | "## " | "### " | "#### " | "##### " | "###### "
```

Damit gilt zugleich:

* nach den `#` muss genau ein Leerzeichen folgen
* vor den `#` dürfen keine Leerzeichen stehen

_Anmerkung:_ Das Leerzeichen ist hier bewusst Teil von `hashes`. Das vereinfacht den Parser: Nach dem Präfix kommt entweder direkt die Nummer, direkt der Anchor oder direkt der Titel.

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

* `number` ist optional.
* Wenn `number` vorkommt, steht sie direkt nach `hashes` und wird von genau einem Leerzeichen gefolgt.
* `anchor` ist optional.
* Wenn `anchor` vorkommt, steht er direkt nach `hashes` oder direkt nach `number SP`.
* Zwischen `</a>` und dem ersten Zeichen des Titels steht **kein** Leerzeichen.
* Innerhalb des Titels bleiben Leerzeichen und Zeichen unverändert erhalten.
* Nur Überschriften, die exakt dieser Positionslogik entsprechen, dürfen von `mdtoc` umgeschrieben werden.

_Erklärung:_

* Die fehlende Leerstelle zwischen `</a>` und Titel bleibt bewusst erhalten, weil sie Teil des verwalteten Render-Formats ist.
* Die Motivation ist jetzt aber nicht mehr `dumeng`-Kompatibilität, sondern ein eindeutig wiedererkennbares und idempotentes Render-Schema.

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

* `### 2024 roadmap` ist **keine** Nummer, weil das erste Token nicht auf `.` endet.
* `### 3D graphics` ist **keine** Nummer, weil das erste Token kein reines `x.y.z.`-Muster ist.
* `### 2.1. API` ist eine verwaltete Nummernsyntax.

_Anmerkung:_ Das Muster `### 2.1. API` ist damit bewusst für `mdtoc` reserviert. Wer eine freie Überschrift exakt in diesem Format schreibt, verwendet dieselbe Syntax wie das Tool.

### 5.4 Unterstützter Markdown-Subset

`mdtoc` ist kein allgemeiner Markdown-Parser.

Für Überschriften gilt in v1:

* nur ATX-Überschriften
* nur die oben definierte Heading-Syntax
* keine Setext-Überschriften
* keine impliziten oder mehrdeutigen Sonderfälle

Der praktische Vorfilter lautet damit mindestens:

```text
^#{1,6} 
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
- title_markup  // Titelbereich wie im Dokument, aber ohne verwaltete Nummer und ohne verwalteten Inline-Anker
- title_text    // Plain-Text-Interpretation von title_markup; Quelle für ToC-Linktext und Anchor-ID
- number        // abgeleitet oder leer
- anchor_id     // abgeleitet oder leer
```

Semantisch wichtig sind nur:

* `level`
* `title_markup`
* `title_text`

Abgeleitet werden daraus:

* `number`
* `anchor_id`

_Erklärung:_

* Die Trennung `title_markup` vs. `title_text` bleibt sinnvoll, auch wenn `mdtoc` die Ableitung von `title_text` nicht selbst spezifiziert.
* `title_text` ist in v1 vollständig an `goldmark` delegiert.
* Maßgeblich ist die von `goldmark` gelieferte Plain-Text-Interpretation des Heading-Inhalts.
* `mdtoc` definiert dafür in v1 keine eigene abweichende Textableitungslogik.

### 6.2 Dokumentzustand

Ein Dokument befindet sich für `mdtoc` praktisch in einem dieser Zustände:

* `unmanaged`  
  Kein gültiger `mdtoc`-Container mit gültigem Config-Block vorhanden.

* `generated`  
  Das Dokument entspricht dem aus Inhalt + Config abgeleiteten Sollzustand.

* `stripped`  
  Die verwalteten Artefakte sind entfernt, der Container mit Config bleibt erhalten.

### 6.3 Verarbeitungspipeline

Die Verarbeitung folgt immer demselben einfachen Muster:

```text
parse -> normalize -> derive -> render
```

Das bedeutet:

* **parse**: Container, Config und Überschriften erkennen
* **normalize**: verwaltete Nummern und verwaltete Anchors semantisch entfernen
* **derive**: Nummern, Anchor-IDs und ToC neu berechnen
* **render**: Dokument deterministisch zurückschreiben

_Anmerkung:_ Das soll keine große AST-Architektur erzwingen. Es legt nur fest, welche Informationen semantisch zählen und welche nur Render-Artefakte sind.

### 6.4 Gültigkeitsbereich von `min-level` und `max-level`

Diese Fassung geht von folgender, leicht verständlicher Regel aus:

* `min-level` und `max-level` filtern dieselbe Menge an Überschriften für
  * ToC-Erzeugung
  * Nummerierung
  * Anchor-Erzeugung

Praktische Folge:

* Bei `generate` werden zunächst alle verwalteten Nummern und verwalteten Anchors aus allen verwalteten Überschriften entfernt.
* Danach werden Nummern und Anchors nur für Überschriften innerhalb des aktiven Level-Bereichs neu gesetzt.
* Überschriften außerhalb des Bereichs bleiben inhaltlich erhalten, werden aber nicht neu verwaltet.

_Verweis:_

* Dieselbe Regel wird in Abschnitt 10 für den ToC noch einmal verwendet.
* Das ist Absicht; beide Stellen beschreiben denselben Vertrag aus zwei Blickwinkeln.

## 7. Config-Block

Der Config-Block hat genau diese Form:

```html
<!-- mdtoc-config
container-version=v2
numbering=true
min-level=2
max-level=4
anchor=github
toc=true
state=generated
-->
```

Regeln:

* Der Config-Block ist zeilenbasiert.
* Jedes Feld steht in einer eigenen Zeile.
* Alle Felder verwenden `key=value`.
* Die Feldreihenfolge ist fest:
  1. `container-version`
  2. `numbering`
  3. `min-level`
  4. `max-level`
  5. `anchor`
  6. `toc`
  7. `bullets`
  8. `state`
* Zulässige Werte:
  * `container-version=v2`
  * `numbering=true|false`
  * `anchor=github|gitlab|off`
  * `toc=true|false`
  * `bullets=auto|*|-|+`
  * `state=generated|stripped`
* `min-level` und `max-level` sind positive ganze Zahlen.
* `min-level` darf nicht größer als `max-level` sein.
* `max-level` darf nicht größer als 6 sein.
* `generate` schreibt alle Generator-Optionen in den Config-Block; nicht angegebene Optionen mit Default-Wert.
* Legacy-Config-Blöcke ohne `container-version` bleiben lesbar und werden als implizites `v1` behandelt.
* Neuer Code zum Schreiben von Config-Blöcken erzeugt `container-version=v2`.
* `--file`, `--help`, `--version`, `--verbose` und `--raw` werden nicht persistiert.
* `strip` behält den Config-Block und setzt nur `state=stripped`.
* `strip --raw` entfernt den Config-Block vollständig.
* `toc=off` bedeutet: der verwaltete ToC-Bereich bleibt Teil des Containers, wird aber leer gerendert.

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

| Option                  | Default | Bedeutung                                                              |
|-------------------------|---------|------------------------------------------------------------------------|
| `--numbering <on\|off>` | `on`    | Kapitelnummern aktivieren oder deaktivieren                            |
| `--min-level <N>`       | `2`     | minimale verwaltete Heading-Ebene (>=1)                                |
| `--max-level <N>`       | `4`     | maximale verwaltete Heading-Ebene (<=6)                                |
| `--anchor <github\|gitlab\|false>` | `github` | Anchor-Profil wählen oder Inline-Anker deaktivieren            |
| `--toc <on\|off>`       | `on`    | rendert den verwalteten ToC-Bereich bei `on`, lässt ihn bei `off` leer |
| `--file <name>`         | –       | Datei lesen und überschreiben                                          |
| `--verbose`             | `off`   | Diagnose- und Ablaufmeldungen auf `stderr`                             |
| `--help`                | –       | Hilfe anzeigen                                                         |

Kurzformen:

| Option        | Kurzform |
|---------------|----------|
| `--numbering` | `-n`     |
| `--anchor`    | `-a`     |
| `--file`      | `-f`     |
| `--verbose`   | `-v`     |
| `--help`      | `-h`     |

### 8.3 I/O- und Logging-Verhalten

* Mit `--file` wird die Datei gelesen und überschrieben.
* Ohne `--file` kommt die Eingabe von `stdin` und die Dokumentausgabe geht auf `stdout`.
* Erfolgreiche Kommandos erzeugen keine Ausgabe, außer bei `--help`, `--version` oder `--verbose`.
* Fehler und Diagnosemeldungen gehen ausschließlich auf `stderr`.

## 9. Kommandos

### 9.1 `generate`

Verhalten:

1. Dokument parsen.
2. Wenn kein verwalteter Container vorhanden ist, kompletten Container am Dateianfang anlegen.
3. Wenn Marker-Struktur oder Config ungültig ist: Fehler und keine Änderung.
4. Vorhandene verwaltete Artefakte semantisch entfernen:
   * ToC-Inhalt
   * verwaltete Kapitelnummern
   * verwaltete Inline-Anker
5. Relevante Überschriften bestimmen.
6. Nummern neu berechnen, falls `numbering=true`.
7. `anchor_id` für alle relevanten Überschriften neu berechnen.
8. Verwaltete Inline-Anker nur rendern, falls `anchor!=off`.
9. ToC neu rendern, falls `toc=true`; andernfalls den verwalteten ToC-Bereich leer rendern.
10. Überschriften neu rendern.
11. Config neu rendern und `state=generated` setzen.
12. Dokument zurückschreiben.

Zusätzliche Regeln:

* Nummerierung und Anchor-ID sind strikt entkoppelt.
* Anchor-IDs werden nur aus dem unnummerierten Titel berechnet.
* Duplicate IDs werden deterministisch aufgelöst.
* Fremder Inhalt im ToC-Bereich wird nicht gelöscht, sondern als HTML-Kommentar erhalten.
* Bei Erfolg ist das Ergebnis idempotent.

Beispiel für eine gerenderte Überschrift:

```md
### 4.1. <a id="open-source"></a>Open source
```

_Erklärung:_

* Zusätzliche benutzerdefinierte Inline-Elemente im Titel sind nicht grundsätzlich verboten.
* Für die normative Ableitung von `anchor_id` zählt aber nicht das rohe Markup, sondern `title_text` gemäß Abschnitt 6 und Abschnitt 11.

### 9.2 `strip`

Verhalten:

* benötigt einen gültigen Config-Block
* entfernt verwalteten ToC-Inhalt
* entfernt verwaltete Kapitelnummern
* entfernt verwaltete Inline-Anker
* behält äußeren Container
* behält Config-Block
* setzt `state=stripped`

Nach `strip` ist damit diese Struktur weiterhin gültig:

```md
<!-- mdtoc -->
<!-- mdtoc-config
numbering=true
min-level=2
max-level=4
anchor=github
toc=true
state=stripped
-->
<!-- /mdtoc -->
```

Fehlerfall:

* kein gültiger Config-Block -> Fehler
* keine implizite Reparatur

### 9.3 `strip --raw`

Verhalten:

* ignoriert den Config-Block
* entfernt den kompletten verwalteten Container, falls vorhanden:
  * `<!-- mdtoc -->`
  * ToC-Inhalt
  * `mdtoc-config`
  * `<!-- /mdtoc -->`
* entfernt zusätzlich verwaltete Kapitelnummern
* entfernt zusätzlich verwaltete Inline-Anker

Konservative Regel:

* Wenn nicht sicher entschieden werden kann, ob eine Nummer oder ein Inline-Anker verwaltet ist, bleibt der Inhalt stehen.
* Dieser Fall soll in jedem Fall als Diagnose ausgegeben werden.

Einsatzfälle:

* beschädigte Config
* Migration
* vollständige Entfernung der `mdtoc`-Verwaltung
* Tests

### 9.4 `check`

Verhalten:

* benötigt einen gültigen Config-Block
* rekonstruiert den Sollzustand aus aktuellem Dokumentinhalt und Config
* vergleicht Sollzustand und Istzustand byte-genau
* liefert `0`, wenn beide identisch sind
* liefert bei Abweichung einen Fehler-Exit-Code

Keine Seiteneffekte:

* `check` verändert das Dokument nie

_Interpretation:_

* `check` muss den Sollzustand abhängig von `state` rekonstruieren.
* Bei `state=generated` entspricht der Sollzustand dem Ergebnis von `generate`; bei `state=stripped` dem Ergebnis von `strip`.

_Anmerkung:_ "byte-genau" klingt formaler als es praktisch ist. Gemeint ist: `check` berechnet denselben Text, den `generate` oder `strip` schreiben würden, und vergleicht genau diesen.

## 10. ToC-Regeln

Der ToC basiert auf allen verwalteten Überschriften innerhalb von `min-level` bis `max-level` inklusive.

Render-Regeln:

* Eine Überschrift erzeugt genau einen ToC-Eintrag.
* Die Hierarchie folgt dem Heading-Level.
* Pro zusätzlicher Ebene relativ zu `min-level` wird um zwei Spaces eingerückt.
* Jeder Eintrag ist ein Markdown-Listenpunkt mit Link.

Beispiel:

```md
* [1. Einleitung](#einleitung)
  * [1.1. API](#api)
```

Anzeige im Linktext:

* bei `numbering=true`: `nummer + titel`
* bei `numbering=false`: nur `titel`

Linkziel:

* grundsätzlich `#` + `anchor_id`
* `anchor_id` wird exakt nach Abschnitt 11 berechnet

Verhalten von `anchor`:

* bei `anchor=github` rendert `mdtoc` zusätzlich einen verwalteten Inline-Anker und berechnet `anchor_id` mit dem GitHub-kompatiblen Profil
* bei `anchor=gitlab` rendert `mdtoc` zusätzlich einen verwalteten Inline-Anker und verwendet die GitLab-Profilauswahl
* bei `anchor=off` rendert `mdtoc` keinen verwalteten Inline-Anker; die ToC-Links bleiben trotzdem `#anchor_id`

_Erklärung:_

* `anchor=off` ist damit ein rendererabhängiger Kompatibilitätsmodus.
* Vollständig portable und rendererunabhängige ToC-Links sind nur mit einem Inline-Anchor-Profil wie `github` oder `gitlab` garantiert.

_Verweis:_

* Die eigentliche Norm für `anchor_id` steht ausschließlich in Abschnitt 11.
* Dieser Abschnitt 10 beschreibt nur die Verwendung der bereits berechneten ID im ToC.

## 11. Gemeinsame Slug- und Anchor-ID-Spezifikation (GitHub-kompatibel)

_Designentscheidung:_ In v1 sind `slug` und `anchor_id` absichtlich dieselbe Zeichenkette.  
Es gibt also nur **eine** normative Spezifikation für beide.

Die ID wird deterministisch aus dem **unnummerierten Plain-Text-Titel (`title_text`)** erzeugt.

### 11.1 Ziel

Die erzeugten Werte sollen sein:

* stabil
* deterministisch
* gut lesbar
* in den dokumentierten Grundregeln GitHub-kompatibel
* sowohl für Inline-Anker als auch ToC-Links identisch

### 11.2 Eingabe für die Ableitung

Für jede verwaltete Überschrift gilt:

```text
slug_source := title_text
anchor_id   := slugify(slug_source)
```

Dabei gilt:

* `title_text` ist **nicht** der rohe Titelstring aus der Zeile.
* `title_text` ist die Plain-Text-Interpretation von `title_markup`.
* Verwaltete Nummer und verwalteter Inline-Anker gehören **nicht** zu `title_text`.

_Erklärung:_

* In v1 wird die Ableitung von `title_text` vollständig an `goldmark` delegiert.
* Maßgeblich ist die von `goldmark` gelieferte Plain-Text-Interpretation des Heading-Inhalts.
* `mdtoc` definiert dafür in v1 keine eigene abweichende Textableitungslogik.
* Nur so bleiben Slug-/Anchor-Bildung, ToC-Linktext und GitHub-ähnliches Verhalten konsistent.

### 11.3 GitHub-kompatible Grundregeln

Die Funktion `slugify` MUSS mindestens diese Schritte ausführen:

1. Eingabe ist `title_text`.
2. Buchstaben werden per Unicode-Lowercasing in Kleinschreibung überführt.
3. Markdown-Formatierungszeichen und Inline-Markup tragen nicht als Literalzeichen zum Slug bei; nur ihr sichtbarer Textinhalt zählt.
4. Unicode-Buchstaben und Unicode-Dezimalziffern bleiben erhalten.
5. Läufe aus Leerraum und Satzzeichen **zwischen** erhaltenen Textteilen werden zu genau einem `-` normalisiert.
6. Führende und folgende Läufe aus Leerraum oder Satzzeichen erzeugen **kein** führendes oder folgendes `-`.
7. Wenn der so berechnete Slug bereits in demselben Dokument existiert, wird `-1`, `-2`, `-3`, ... angehängt.

_Interpretation:_

* Diese Regeln folgen den von GitHub dokumentierten Grundregeln in einer für `mdtoc` explizit testbaren Form.
* Für nicht dokumentierte Randfälle macht `mdtoc` in den folgenden Unterpunkten weitere Festlegungen.

### 11.4 Explizite Festlegungen für Randfälle

Zusätzlich gilt in `mdtoc` v1:

* Symbole, Emojis und sonstige Nicht-Buchstaben-/Nicht-Ziffern-Zeichen werden entfernt.
* Läufe aus Leerraum/Satzzeichen werden nicht mehrfach als `--`, `---` usw. abgebildet, sondern zu genau einem `-` zusammengezogen.
* Die Kollisionsauflösung beginnt beim **zweiten** Vorkommen mit `-1`.
* Wenn der normalisierte Slug leer wird, verwendet `mdtoc` den Fallback `section`.
* Weitere Kollisionen auf diesem Fallback werden mit `section-1`, `section-2`, ... aufgelöst.

_Erklärung:_

* Der Fallback `section` ist eine bewusste `mdtoc`-Festlegung.
* GitHubs öffentliche Grundregeln beschreiben diesen Leerslug-Randfall nicht explizit.

### 11.5 Beziehung zur Inline-Anker-Syntax

Wenn `anchor!=off`, rendert `mdtoc` exakt diese Form:

```html
<a id="anchor_id"></a>
```

Dabei gilt:

* der String in `id="..."` MUSS exakt dem nach diesem Abschnitt berechneten `anchor_id` entsprechen
* `slug`, `anchor_id` und ToC-Linkziel sind damit dieselbe Zeichenkette

_Verweis:_

* Abschnitt 5 definiert nur die Position und das Render-Format des Inline-Ankers.
* Die Zeichenkette innerhalb von `id="..."` wird ausschließlich hier in Abschnitt 11 normiert.

### 11.6 Beispiele

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
### This'll be a _Helpful_ Section About the Greek Letter Θ!
```

→

```text
thisll-be-a-helpful-section-about-the-greek-letter-θ
```

#### Beispiel 3

```md
### Übergrößenträger & naïve façade – déjà vu!
```

→

```text
übergrößenträger-naïve-façade-déjà-vu
```

#### Beispiel 4

```md
### 中文 русский عربى
```

→

```text
中文-русский-عربى
```

#### Beispiel 5

```md
### 🚀 !!! 
```

→

```text
section
```

#### Beispiel 6

Zwei identische Überschriften `### API` ergeben:

```text
api
api-1
```

## 12. Fehlerverhalten, Logging und Exit-Codes

Fehlerfälle:

* fehlender oder unvollständiger `mdtoc`-Container
* fehlender Config-Block bei `strip` oder `check`
* ungültiger Config-Block
* Parsing-Fehler
* ungültige Optionen

Grundregeln:

* Fehler werden auf `stderr` ausgegeben.
* Bei Fehlern gibt es keine implizite Reparatur, außer dem explizit erlaubten Anlegen eines neuen Containers durch `generate`, wenn noch gar keine `mdtoc`-Verwaltung existiert.
* Erfolgreiche Kommandos schreiben keine Statusmeldungen auf `stdout`.

Empfohlene Exit-Codes:

* `0` -> Erfolg
* `1` -> Parsing-, Config- oder CLI-Fehler
* `2` -> `check` hat eine Abweichung gefunden

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

_Verweis:_

* Die Idempotenz ist bereits in Abschnitt 1 als Grundprinzip und in Abschnitt 9 als Kommandosemantik angelegt.
* Dieser Abschnitt 13 wiederholt den Vertrag absichtlich noch einmal in Testform.

## 14. Erweiterbarkeit

Mögliche spätere Erweiterungen:

* alternative Anchor-Styles
* alternative ToC-Formate
* Versionierung im Config-Block
* weitere Ausgabeformate

_Anmerkung:_ Diese Punkte sind ausdrücklich Erweiterungen. Sie sollen v1 nicht unnötig komplex machen.

## 15. Empfohlene Go-Implementierungsbasis: `goldmark` (informativ)

Für eine Go-Implementierung ist die Verwendung von `goldmark` sinnvoll.

Empfohlenes Setup:

* Modul: `github.com/yuin/goldmark`
* Zielbereich: aktuelle stabile `1.x`-Version
* konkrete Referenz für diese Spezifikation: `v1.8.x`
* empfohlene Erweiterung: `extension.GFM`

Empfehlung für `mdtoc`:

* `goldmark` SOLL für Parsing, Heading-Erkennung, Ableitung von `title_text` aus Überschriften und robuste Behandlung von Fenced Code Blocks verwendet werden.
* Die Ableitung von `title_text` wird in v1 vollständig an `goldmark` delegiert.
* Maßgeblich ist die von `goldmark` gelieferte Plain-Text-Interpretation des Heading-Inhalts.
* `mdtoc` definiert dafür in v1 keine eigene abweichende Textableitungslogik.
* Die normative Slug-/Anchor-ID aus Abschnitt 11 SOLL weiterhin von `mdtoc` selbst gemäß dieser Spezifikation berechnet werden.
* `parser.WithAutoHeadingID()` SOLL daher nicht die normative Quelle der `anchor_id` sein.
* Wenn im Code trotzdem eine `goldmark`-eigene ID-Erzeugung genutzt werden soll, MUSS dafür eine eigene `parser.IDs`-Implementierung verwendet werden, die Abschnitt 11 exakt einhält.

_Erklärung:_

* `goldmark` reduziert den Implementationsaufwand deutlich, weil Heading-Struktur, Fences, Inline-Markup und Source-Positionen nicht per Regex nachgebaut werden müssen.
* Die eigentliche Fachlogik von `mdtoc` bleibt trotzdem klein und eigenständig: Container finden, verwaltete Überschriften normalisieren, Nummern/IDs ableiten, deterministisch rendern.

---
