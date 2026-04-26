> Anweisung: Alle aktuell mit ">" beginneneden Zeilen kommen von mir und sind einzuarbeiten und dann zu entfernen. Deine neuen Kommentare und Fragen mit ">" einleiten.

# Issue #54: Robustere Analyse und Behandlung fehlerhafter `mdtoc`-Container

## Ziel

`mdtoc` soll Dokumente mit vorhandenem oder teilweise beschädigtem `mdtoc`-Container robuster analysieren und konsistenter behandeln. Das betrifft insbesondere Fälle, in denen der äußere Container formal vorhanden ist, der Config-Block jedoch beschädigt, inkonsistent, falsch positioniert oder teilweise außerhalb des Containers liegt.

Das Ziel ist nicht, beliebige kaputte Markdown-Dateien heuristisch zu "reparieren". Ziel ist eine klar definierte, testbare und für Anwender nachvollziehbare Behandlung von Dokumenten, die erkennbar einen `mdtoc`-Container enthalten oder enthalten sollten.

## Problem

Aktuell ist das Verhalten zwischen den Kommandos nicht ausreichend vereinheitlicht:

* `check` soll nur reporten und niemals schreiben.
* `regen` und `strip` sollen nur dann aktiv schreiben, wenn der Config-Block konsistent und verwertbar ist.
* `generate` soll robuster mit bereits vorhandenem Container-Zustand umgehen können, insbesondere wenn generierte Reste oder beschädigte Config-Strukturen im Dokument liegen.
* `strip --raw` soll generierte Inhalte möglichst vollständig entfernen, solange keine klar definierte Abbruchbedingung vorliegt.

Dafür braucht `mdtoc` vor jeder eigentlichen Aktion eine gemeinsame Container-Analyse mit eindeutig benannten Zuständen.

## Begriffe

### Excluded Regions

Excluded Regions sind Bereiche, die von der Container- und Config-Analyse im ersten Lauf ignoriert werden müssen. Dazu gehören:

* fenced code blocks
* ausgeschlossene `mdtoc off` / `mdtoc on`-Bereiche
* andere bereits definierte ignorierte Regionen gemäß aktueller Parser-Logik

Wichtig ist: Alle Erkennungen von Container-Markern, Config-Block-Strukturen und relevanten Zeilen müssen zunächst nur außerhalb solcher Regionen stattfinden.

### Container

Ein Container ist die äußere, von `mdtoc` verwaltete Struktur:

```md
<!-- mdtoc -->
[ToC-Bereich]
<!-- mdtoc-config
...
-->
<!-- /mdtoc -->
```

Zum Container gehören:

* Startmarker
* ToC-Bereich
* Config-Block
* Endmarker
* gegebenenfalls zusätzliche Zeilen, die versehentlich innerhalb dieses Bereichs gelandet sind

### Intakte äußere Container-Struktur

Die äußere Container-Struktur ist intakt, wenn:

* genau ein Startmarker erkannt wird
* genau ein Endmarker erkannt wird
* beide Marker außerhalb von Excluded Regions liegen
* der Startmarker vor dem Endmarker liegt

Diese Aussage betrifft zunächst nur die äußere Klammerung, noch nicht die Gültigkeit oder Position des Config-Blocks.

### ToC-Bereich

Der ToC-Bereich ist der Bereich zwischen Container-Start und Beginn des Config-Blocks.

Er kann enthalten:

* eindeutig generierte ToC-Zeilen
* zusätzliche Zeilen, die nicht von `mdtoc` generiert wurden, aber im Container gelandet sind

### Eindeutig generierte ToC-Zeilen

Eine ToC-Zeile gilt als eindeutig generiert, wenn sie syntaktisch einer von `mdtoc` erzeugten ToC-Zeile entspricht. Beispiel:

```text
  * [2.1. Releases](#releases)
```

Praktische Anforderungen an die Erkennung:

* führende und abschließende Whitespaces dürfen toleriert werden
* es muss ein Bullet folgen
* danach muss eine `[]()`-Struktur folgen
* der Link im `()`-Teil muss zum Slug des Textes im `[]`-Teil passen
* wenn kein intakter Config-Block vorliegt, ist für diese Plausibilisierung standardmäßig das GitHub-Slug-Verhalten anzunehmen

Die Toleranz für Whitespaces am Zeilenende ist wichtig, damit unsichtbare Editorreste nicht fälschlich als fremde Zeilen behandelt werden.

### Config-Block

Der Config-Block ist der von `mdtoc` verwaltete Konfigurationsbereich innerhalb des Containers.

Er besteht aus:

* Config-Startmarker
* Config-Zeilen
* Config-Endmarker
* gegebenenfalls zusätzlichen Zeilen, die versehentlich innerhalb des Blocks gelandet sind

### Config-Zeilen

Config-Zeilen sind reine `key=value`-Zeilen mit optionalen Whitespaces am Zeilenende.

### Intakte äußere Config-Block-Struktur

Die äußere Config-Block-Struktur ist intakt, wenn:

* genau ein Config-Start erkannt wird
* genau ein Config-Ende erkannt wird
* beide Marker außerhalb von Excluded Regions liegen
* der Config-Start vor dem Config-Ende liegt

Diese Aussage betrifft noch nicht:

* ob der Block im Container liegt
* ob die Zeilenanzahl korrekt ist
* ob die erwarteten Schlüssel vorhanden sind
* ob die Werte gültig sind

### Konsistenter Config-Block

Ein Config-Block ist konsistent, wenn zusätzlich zur intakten äußeren Config-Block-Struktur gilt:

* der Block liegt vollständig innerhalb des Containers
* der Block steht an der dafür vorgesehenen Position innerhalb des Containers
* die erwarteten Schlüssel sind vorhanden
* die Reihenfolge entspricht der definierten Struktur
* die Anzahl der relevanten Zeilen passt
* alle Werte sind gültig

## Grundprinzip für alle Kommandos

Alle Kommandos müssen zunächst denselben Analyse-Lauf durchführen und dabei den Container-Zustand klassifizieren, bevor sie schreiben, vergleichen oder abbrechen.

Dieser erste Analyse-Lauf muss:

* Excluded Regions überspringen
* die äußere Container-Struktur bewerten
* die äußere Config-Block-Struktur bewerten
* die Lage des Config-Blocks relativ zum Container bewerten
* den Inhalt des ToC-Bereichs grob in generierte und nicht generierte Zeilen einordnen
* den Config-Inhalt auf Konsistenz prüfen, sofern ein formal verwertbarer Block vorliegt

## Soll-Verhalten der Kommandos

### `check`

`check` verändert niemals Dokumentinhalte.

`check` darf:

* den Ist-Zustand analysieren
* Inkonsistenzen reporten
* mit Fehlerstatus abbrechen

`check` darf nicht:

* Zeilen löschen
* Zeilen verschieben
* Config oder ToC neu schreiben

### `regen`

`regen` darf nur dann schreiben, wenn ein konsistenter, verwertbarer Config-Block vorliegt.

Wenn der Config-Block nicht konsistent ist, muss `regen` reporten und abbrechen.

### `strip`

`strip` darf nur dann schreiben, wenn ein konsistenter, verwertbarer Config-Block vorliegt.

Wenn der Config-Block nicht konsistent ist, muss `strip` reporten und abbrechen.

### `generate`

`generate` ist das toleranteste schreibende Kommando. Es soll versuchen, vorhandene Struktur soweit sinnvoll auszuwerten und anschließend den Zielzustand vollständig neu aufzubauen.

Konzeptionell soll `generate` so gedacht werden:

1. vorhandene Container-Informationen analysieren
2. falls ein intakter Config-Block vorliegt, dessen Werte mit CLI-Werten mergen, wobei CLI-Werte Vorrang haben
3. vorhandenen generierten Inhalt entfernen
4. Zielzustand vollständig neu rendern

Praktisch darf die interne Implementierung anders aussehen, solange das externe Verhalten diesem Modell entspricht und bestehende, bereits korrekte Tests nicht unnötig destabilisiert werden.

### `strip --raw`

`strip --raw` entfernt allen generierten Inhalt, sofern keine definierte Abbruchbedingung greift.

Dabei gilt:

* generierte ToC-Zeilen sollen entfernt werden
* generierte Config-Zeilen sollen entfernt werden
* zusätzliche, nicht generierte Zeilen im Container sollen erhalten bleiben
* verwaltete Heading-Artefakte wie Nummerierung und generierte Inline-Anker sollen entfernt werden

`strip --raw` ist ausdrücklich der toleranteste Bereinigungsmodus.

## Zu behandelnde Container-Zustände

Die Analyse muss mindestens die folgenden Fälle unterscheiden.

### 1. Äußere Container-Struktur defekt

Beispiele:

* nur Startmarker vorhanden
* nur Endmarker vorhanden
* Startmarker nach Endmarker
* doppelte äußere Marker

Erwartetes Verhalten:

* Abbruch
* Meldung mit Zeilennummern bzw. betroffenen Markerpositionen
* keine schreibende Aktion außer dort, wo ein ausdrücklich definierter Fallback für `strip --raw` erlaubt ist

### 2. Excluded Region innerhalb des Containers

Wenn innerhalb des erkannten Containers Excluded Regions vorkommen, soll dies nicht stillschweigend akzeptiert werden.

Erwartetes Verhalten:

* Abbruch
* Meldung mit Zeilennummern

Begründung:

* Ein verwalteter Bereich muss strukturell klar und vollständig analysierbar bleiben.
* Ignorierte Regionen innerhalb des Containers machen die Zuordnung zwischen generiertem und fremdem Inhalt unnötig unsicher.

### 3. Äußere Config-Block-Struktur defekt

Beispiele:

* Config-Start ohne Config-Ende
* Config-Ende ohne Config-Start
* doppelter Config-Start
* doppeltes Config-Ende
* falsche Reihenfolge

Erwartetes Verhalten:

* Abbruch mit Zeilennummerninfo
* `regen` und `strip` schreiben nicht
* `check` reportet nur
* `generate` und `strip --raw` dürfen nur dann tolerant fortfahren, wenn das definierte Bereinigungsmodell dies erlaubt

### 4. Config-Block außerhalb des Containers

Hier ist die äußere Config-Block-Struktur zwar grundsätzlich erkennbar, der Block liegt aber ganz oder teilweise außerhalb des Containers.

Erwartetes Verhalten:

* die generierten Config-Zeilen dieses Blocks gelten als entfernbar
* nicht generierte Zusatzzeilen sollen nicht pauschal gelöscht werden
* das Verhalten muss für `generate` und `strip --raw` klar definiert und testbar sein
* `regen`, `strip` und `check` sollen diesen Zustand reporten statt von einem gültigen Managed State auszugehen

### 5. Config-Block-Struktur äußerlich intakt, aber inhaltlich inkonsistent

Beispiele:

* falsche Zeilenanzahl
* unerwartete Schlüssel
* falsche Reihenfolge
* ungültige Werte
* zusätzliche generierte Config-Zeilen oder beschädigte generierte Config-Zeilen

Erwartetes Verhalten:

* generierte Config-Zeilen sollen als entfernbar gelten
* `regen` und `strip` brechen reportend ab
* `check` reportet nur
* `generate` und `strip --raw` dürfen den Zielzustand aus einer bereinigten Sicht neu aufbauen

## Gewünschte Normalform der Verarbeitung

Für die Implementierung ist eine Trennung in zwei Ebenen sinnvoll:

### Ebene 1: Struktur-Scan

Ein toleranter, rein struktureller Scan erkennt:

* Container-Marker
* Config-Marker
* ihre Positionen
* Kollisionen mit Excluded Regions
* grob generierte ToC-Zeilen
* grob generierte Config-Zeilen

Dieser Scan darf noch nicht voraussetzen, dass der Config-Block bereits vollständig parsebar ist.

### Ebene 2: Semantische Validierung

Wenn die Struktur dafür geeignet ist, folgt die semantische Validierung:

* Config parsen
* Werte validieren
* Konsistenzzustand ableiten
* erlaubte Aktion für das jeweilige Kommando bestimmen

Diese Zweiteilung reduziert Sonderfälle und erlaubt einheitliche Entscheidungen für alle Kommandos.

## Erwartete Wirkung auf die Kommandos

Die Kommandos sollen nach derselben Analyse zu unterschiedlichen Aktionsrechten kommen:

* `check`: analysieren und reporten
* `regen`: nur bei konsistentem Config-Block schreiben
* `strip`: nur bei konsistentem Config-Block schreiben
* `generate`: bereinigen und neu aufbauen
* `strip --raw`: maximal tolerant bereinigen

> Frage: Was bedeutet die Unterscheidung zwischen "bereinigen" und "maximal tolerant bereinigen hier? Ich vermute der Bereinigungslevel bei generate und strip --raw sollte gleich sein, oder? Anpassen.

Dadurch wird vermieden, dass jedes Kommando eigene heuristische Sonderfälle pflegt.

## Abgrenzung

Diese Issue verlangt nicht:

* automatische Reparatur beliebiger fremder Markdown-Strukturen
* stillschweigendes Umschreiben unklarer oder mehrdeutiger Bereiche
* Lockerung der Managed-Format-Definition im regulären Erfolgsfall

Diese Issue verlangt:

* klar benannte Zustände
* wohldefinierte Reaktionen je Kommando
* nachvollziehbare Fehlermeldungen
* testbare Regeln für tolerante Bereinigung

## Vorschlag für Akzeptanzkriterien

Die Issue ist erst dann abgeschlossen, wenn mindestens die folgenden Punkte erfüllt sind.

### Analyse und Modell

* Es gibt eine gemeinsame Container-Analyse vor der eigentlichen Aktion.
* Diese Analyse überspringt Excluded Regions im ersten Lauf.
* Die Analyse unterscheidet äußere Strukturfehler, Lagefehler und inhaltliche Inkonsistenzen.

### Kommandoverhalten

* `check` schreibt nie.
* `regen` schreibt nur bei konsistentem Config-Block.
* `strip` schreibt nur bei konsistentem Config-Block.
* `generate` kann verwertbare Altzustände toleranter bereinigen und den Zielzustand vollständig neu erzeugen.
* `strip --raw` entfernt generierte Inhalte auch dann, wenn der Config-Block nicht mehr vollständig parsebar ist, solange keine definierte Abbruchbedingung greift.

### Erhaltung nicht generierter Inhalte

* Zusätzliche Zeilen im Container bleiben erhalten, sofern sie nicht eindeutig generiert sind.
* Zusätzliche Zeilen im Config-Block werden nicht pauschal als frei löschbar behandelt, außer soweit sie eindeutig als generierte Config-Zeilen klassifizierbar sind oder die konkrete Regel dies erlaubt.

> Hinweis: Die zusätzlichen Zeilen sollen räumlich in der Nähe und gleicher Reihenfolge erhlten bleiben, am einfachsten wohl direkt nach dem Container. Einarbeiten.

### Fehlermeldungen

* Strukturfehler werden mit brauchbaren Positionsinformationen gemeldet.
* Der Nutzer kann aus der Meldung erkennen, ob ein Strukturproblem, ein Lageproblem oder eine inhaltliche Inkonsistenz vorliegt.

### Tests

Für jeden relevanten Zustand gibt es explizite Tests mindestens für:

* `check`
* `regen`
* `strip`
* `generate`
* `strip --raw`

Zusätzlich sollen Tests enthalten:

* Varianten mit Excluded Regions
* Config-Block außerhalb des Containers
* beschädigte Config-Block-Grenzen
* inkonsistente Config-Inhalte
* zusätzliche nicht generierte Zeilen im ToC-Bereich
* zusätzliche nicht generierte Zeilen im Config-Kontext, soweit deren Behandlung definiert ist

## Umsetzungshinweis

Die Umsetzung erscheint grundsätzlich machbar, wenn sie als Erweiterung der bestehenden Parser- und Fallback-Logik erfolgt und nicht als vollständiger Neuansatz. Besonders naheliegend ist:

* den heutigen strikten Parse-Pfad für gültige Dokumente beizubehalten
* davor oder daneben einen toleranteren Struktur-Scan einzuführen
* die Aktionsentscheidung pro Kommando auf einen gemeinsamen Analyse-Report zu stützen

So kann die bestehende, bereits getestete Logik für den Normalfall weitgehend stabil bleiben, während die Sonderfälle systematischer behandelt werden.

## Offene Präzisierungen vor der Umsetzung

Folgende Punkte sollten vor oder während der Umsetzung noch endgültig festgezogen werden:

* Welche Excluded Regions gelten exakt für die Container-Analyse, und sollen sie vollständig mit den bisherigen ignorierten Regionen übereinstimmen?

> Die Begriffe "excluded" und "ignored" müssen in diesem Kontext sauber definiert werden. Code-Fences sind ignored und excluded sind geklammert, richtig? Also sollte "excluded und ignored" geschrieben werden, oder? Ja, und kein Unterschied in der bisherigen Bedeutung.

* Soll ein Config-Block außerhalb des Containers nur reportet oder in toleranten Modi aktiv bereinigt werden, wenn daneben noch ein regulärer Container existiert?

> aktiv bereinigen unter Beibehaltung nicht-generierter Zeilen.

* Wie weit darf `generate` bei beschädigter Struktur gehen, bevor ebenfalls zwingend abgebrochen wird?

> Ich halte gleiche Regeln für für strip --raw für sinnvoll.

* Sollen zusätzliche Zeilen innerhalb eines Config-Blocks grundsätzlich erhalten bleiben, wenn sie nicht eindeutig generiert sind, oder soll bereits die Existenz solcher Zeilen den gesamten Block als unbrauchbar markieren?

> Die Zeilen sollen grundsätzlich erhalten bleiben, aber der generierte Block verschwinden und neu aufgebaut werden.

* Welche konkreten Abbruchbedingungen gelten für `strip --raw`, damit der Modus tolerant bleibt, aber nicht in mehrdeutigen Strukturen destruktiv wird?

> strip --raw bricht, wie besprochen bei defekter äußerer Container Struktur ab mit einer Liste Zeile:Inhalt. Diese bereits erarbeitete Festlegung vermisse ich in diesem Dokument.

## Bedeutung für die Spezifikation

Die hier definierte Logik ist nicht nur Implementierungsdetail. Sie sollte nach erfolgreicher Umsetzung auch in die `mdtoc`-Spezifikation überführt werden, insbesondere für:

* Container-Zustände
* Kommandoverhalten bei inkonsistenten Managed-Strukturen
* tolerante Bereinigung durch `strip --raw`
* Verhältnis zwischen strukturellem Scan und semantischer Validierung
