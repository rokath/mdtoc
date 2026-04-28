
# Issue #54: Robustere Analyse und Behandlung fehlerhafter `mdtoc`-Container (GPT Chatverlauf)

>## User
>
> Das ist der Text einer Issue,  die kontrolliert in sub issues umgesetzt werden soll. Analysiere und mach einen Vorschlag für mehrere testbare sub issues um die Aufgabe umfassend zu lösen

---
## Ziel

`mdtoc` soll Dokumente mit vorhandenem oder teilweise beschädigtem `mdtoc`-Container robuster analysieren und konsistenter behandeln. Das betrifft insbesondere Fälle, in denen der äußere Container formal vorhanden ist, der Config-Block jedoch beschädigt, inkonsistent, falsch positioniert oder teilweise außerhalb des Containers liegt.

Das Ziel ist nicht, beliebige kaputte Markdown-Dateien heuristisch zu "reparieren". Ziel ist eine klar definierte, testbare und für Anwender nachvollziehbare Behandlung von Dokumenten, die erkennbar einen `mdtoc`-Container enthalten oder enthalten sollten.

Ein Hauptziel ist: Die Implementierung darf niemals riskieren, möglichen Nutzerinhalt zu löschen, der versehentlich irgendwo innerhalb des Containers geschrieben wurde, einschließlich innerhalb eines beschädigten Config-Blocks.

Daraus folgen zwei weitere Leitlinien:

* Möglicher Nutzerinhalt darf während Wiederherstellung oder Bereinigung nicht stillschweigend verloren gehen.
* Das Verhalten soll einfach, deterministisch und leicht erklärbar bleiben.

Das bedeutet:

* sicher als generiert erkannte ToC-Zeilen dürfen verworfen werden
* sicher als generiert erkannte Config-Zeilen dürfen verworfen werden
* Leerzeilen dürfen normalisiert oder verworfen werden
* alle anderen Zeilen innerhalb eines beschädigten Containers müssen die Wiederherstellung überleben

Wenn nicht generierter Inhalt irgendwo innerhalb eines beschädigten Containers gefunden wird, soll `generate`:

* einen neuen normalisierten Container aufbauen
* den erhaltenen nicht generierten Inhalt außerhalb des neu aufgebauten Containers platzieren
* die erhaltenen Zeilen in ihrer ursprünglichen Reihenfolge belassen
* eine Warnung ausgeben, die diesen Vorgang erklärt

So wird stiller Datenverlust vermieden, während der neue Container sauber bleibt.

## Problem

Aktuell ist das Verhalten zwischen den Kommandos nicht ausreichend vereinheitlicht:

* `check` soll nur reporten und niemals schreiben.
* `regen` und `strip` sollen nur dann aktiv schreiben, wenn der Config-Block vorhanden, konsistent und verwertbar ist und der Container intakt ist.
* `generate` soll robuster mit bereits vorhandenem Container-Zustand umgehen können, insbesondere wenn generierte Reste und/oder beschädigte Container- und/oder Config-Strukturen im Dokument liegen.
* `strip --raw` soll generierte Inhalte möglichst vollständig entfernen, solange keine klar definierte Abbruchbedingung vorliegt.

Dafür braucht `mdtoc` vor jeder eigentlichen Aktion eine gemeinsame Container-Analyse mit eindeutig benannten Zuständen.

Diese Analyse bzw. Integritätsprüfung soll als gemeinsame Funktion implementiert und von allen Unterkommandos verpflichtend verwendet werden.

## Begriffe

### Ignorierte und ausgeschlossene Bereiche

In diesem Kontext werden die bisher verwendeten Begriffe `ignored` und `excluded` nicht unterschiedlich verwendet, sondern gemeinsam behandelt. Gemeint sind alle Bereiche, die von der Container- und Config-Analyse im ersten Lauf ausgespart werden müssen.

Dazu gehören mindestens:

* Fence-Codeblöcke
* ausgeschlossene `mdtoc off` / `mdtoc on`-Bereiche
* andere bereits definierte ignorierte Bereiche gemäß aktueller Parser-Logik

Für diese Issue gilt daher:

* Code-Fences sind `ignored`
* `mdtoc off` / `mdtoc on`-Bereiche sind `excluded`
* für die Analyse-Regeln dieser Issue werden beide Klassen gemeinsam als auszusparende Bereiche behandelt

Wichtig ist:

* Alle Erkennungen von Container-Markern, Config-Block-Strukturen und relevanten Zeilen müssen zunächst nur außerhalb dieser ignorierten bzw. ausgeschlossenen Bereiche stattfinden.
* Sind äußere Container-Strukturen sicher erkannt, ist zu prüfen, ob innerhalb der festgestellten Containergrenzen ausgeschlossene oder ignorierte Bereiche existieren. Falls ja, sind diese unabhängig von ihrem Inhalt vollständig, also mit Markern, als nicht generierter Inhalt zu behandeln und dürfen daher nicht gelöscht werden.

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

* Startmarker (`<!-- mdtoc -->`)
* ToC-Bereich
* Config-Block (`<!-- mdtoc-config ... -->`)
* Endmarker (`<!-- /mdtoc -->`)
* gegebenenfalls zusätzliche Zeilen, die versehentlich innerhalb dieses Bereichs gelandet sind

Diese zusätzlichen Zeilen bewirken zwar, dass der Container inhaltlich nicht intakt ist, aber die äußere Container-Struktur besteht weiterhin.

### Intakte äußere Container-Struktur

Die äußere Container-Struktur ist intakt, wenn:

* genau ein Startmarker erkannt wird
* genau ein Endmarker erkannt wird
* beide Marker außerhalb von Excluded Regions liegen
* der Startmarker vor dem Endmarker liegt
* zusätzliche eindeutig nicht generierte Zeilen die äußere Markerstruktur nicht mehrdeutig machen

Zusätzliche nicht generierte Zeilen zerstören also nicht den Status "äußerlich intakt", wohl aber gegebenenfalls den Status "inhaltlich intakt".

Diese Aussage betrifft zunächst nur die äußere Klammerung, noch nicht die Gültigkeit oder Position des Config-Blocks.

Wenn keine intakte äußere Container-Struktur eindeutig erkennbar ist, ist dies eine harte Abbruchbedingung. Die Fehlermeldung soll Zeilennummern und Zeileninhalt ausgeben.

Beispiel:

```text
Mehrdeutige mdtoc-Containermarker; bitte manuell korrigieren oder entfernen:
1: <!-- mdtoc -->
17: <!-- /mdtoc -->
42: <!-- /mdtoc -->
```

### Intakter Container

Ein Container gilt nur dann als intakt, wenn zusätzlich zur intakten äußeren Container-Struktur gilt:

* sein ToC-Bereich vollständig und zuverlässig als generierter Inhalt identifizierbar ist
* sein Config-Kontext vollständig und zuverlässig als generierter Inhalt identifizierbar ist
* keine zusätzlichen nicht generierten Zeilen innerhalb des Containers liegen

Die Existenz eines generierten ToC ist dabei für sich genommen kein Intaktheitskriterium, weil der ToC inhaltlich veraltet sein kann. Maßgeblich ist nicht die Aktualität des ToC, sondern die eindeutige Identifizierbarkeit des Container-Inhalts als von `mdtoc` generierter Inhalt.

Ein Container kann also äußerlich intakt sein, inhaltlich aber nicht. Wenn der Container äußerlich und inhaltlich intakt ist, gilt er kurz als intakt. Ein Container, der äußerlich nicht intakt ist, ist eine harte Abbruchbedingung mit Zeilenausgabe im Format `Zeile:Inhalt`, die die relevanten Marker aufführt.

### ToC-Bereich

Der ToC-Bereich ist der Bereich zwischen Container-Start und Beginn des Config-Blocks.

Er kann enthalten:

* eindeutig generierte ToC-Zeilen
* zusätzliche Zeilen, die nicht von `mdtoc` generiert wurden, aber im Container gelandet sind

Solche zusätzlichen nicht generierten Zeilen bewirken, dass der Container als nicht intakt beurteilt wird.

### Eindeutig generierte ToC-Zeilen

Eine ToC-Zeile gilt als eindeutig generiert, wenn sie syntaktisch einer von `mdtoc` erzeugten ToC-Zeile entspricht. Beispiel:

```text
  * [2.1. Releases](#releases)
```

Praktische Anforderungen an die Erkennung:

* führende und abschließende Whitespaces dürfen toleriert werden
* es muss ein Bullet folgen
* danach muss eine `[]()`-Struktur folgen
* der Link im `()`-Teil muss zum Slug des Textes im `[]`-Teil passen; dabei muss auch der Fall doppelter Überschriften berücksichtigt werden
* wenn kein intakter Config-Block vorliegt, ist für diese Plausibilisierung standardmäßig das GitHub-Slug-Verhalten anzunehmen

Die Toleranz für Whitespaces am Zeilenende ist wichtig, damit unsichtbare Editorreste nicht fälschlich als fremde Zeilen behandelt werden.
Es wird nicht verlangt, dass der Text der ToC-Zeile noch zu einer aktuell vorhandenen Überschrift im Markdown-Dokument passt; der Dokumentinhalt kann zwischenzeitlich verändert worden sein.

### Config-Block

Der Config-Block ist der von `mdtoc` verwaltete Konfigurationsbereich innerhalb des Containers.

Er besteht aus:

* Config-Startmarker (`<!-- mdtoc-config`)
* Config-Zeilen
* Config-Endmarker (`-->`)
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

### Zusätzliche Zeilen

Zusätzliche Zeilen sind Zeilen, die innerhalb des Containers oder innerhalb des Config-Kontexts liegen, aber weder eindeutig generierte ToC-Zeilen noch eindeutig generierte Config-Zeilen sind.

Für diese Zeilen gilt:

* sie sollen grundsätzlich erhalten bleiben, außer sie sind leer oder enthalten nur Whitespace
* ihre Reihenfolge soll erhalten bleiben
* sie sollen räumlich in der Nähe des Containers erhalten bleiben
* sie dürfen nicht innerhalb eines als intakt behandelten Containers verbleiben
* bei toleranter Bereinigung sollen sie aus dem Container herausgenommen und standardmäßig direkt nach dem Container wieder ausgegeben werden

Die Gefahr, dass nicht generierte Zeilen im Config-Block gelandet sind, ist geringer als im ToC-Bereich, da der Config-Block im gerenderten Output unsichtbar ist und viele moderne Editoren HTML-auskommentierte Bereiche anders einfärben.

## Grundprinzip für alle Kommandos

Alle Kommandos müssen zunächst denselben Analyse-Lauf durchführen und dabei den Container-Zustand klassifizieren, bevor sie schreiben, vergleichen oder abbrechen.

Diese Validierung ist nicht optional und darf nicht je Kommando unterschiedlich umgangen oder abgeschwächt werden.

Dieser erste Analyse-Lauf muss:

* Excluded Regions überspringen
* die äußere Container-Struktur bewerten
* die Intaktheit des Containers bewerten
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

`regen` darf nur dann schreiben, wenn ein konsistenter, verwertbarer Config-Block vorliegt und der Container als intakt validiert wurde.

Wenn der Config-Block nicht konsistent ist oder der Container nicht intakt ist, muss `regen` reporten und abbrechen.

### `strip`

`strip` darf nur dann schreiben, wenn ein konsistenter, verwertbarer Config-Block vorliegt und der Container als intakt validiert wurde.

Wenn der Config-Block nicht konsistent ist oder der Container nicht intakt ist, muss `strip` reporten und abbrechen.

### `generate`

`generate` ist ein tolerant schreibendes Kommando. Es soll versuchen, vorhandene Struktur soweit sinnvoll auszuwerten und anschließend den Zielzustand vollständig neu aufzubauen.

Konzeptionell soll `generate` so gedacht werden:

1. vorhandene Container-Informationen analysieren
2. falls ein intakter Config-Block vorliegt, dessen Werte mit CLI-Werten mergen, wobei CLI-Werte Vorrang haben
3. falls der vorhandene Zustand nicht intakt oder inkonsistent ist, gedanklich zunächst denselben Bereinigungsschritt wie `strip --raw` anwenden, um den Ausgangszustand zu normalisieren
4. Zielzustand vollständig neu rendern

Praktisch darf die interne Implementierung anders aussehen, solange das externe Verhalten diesem Modell entspricht und bestehende, bereits korrekte Tests nicht unnötig destabilisiert werden.

Für beschädigte Managed-Strukturen soll `generate` dieselben strukturellen Toleranz- und Abbruchregeln wie `strip --raw` verwenden. Der Unterschied zwischen beiden Kommandos liegt nicht im Bereinigungslevel, sondern nur im Zielzustand:

* `generate` bereinigt zunächst mit denselben Strukturregeln wie `strip --raw` und baut anschließend einen neuen gültigen Managed-Zustand auf
* `strip --raw` bereinigt mit denselben Strukturregeln und entfernt anschließend den Managed-Zustand vollständig

### `strip --raw`

`strip --raw` entfernt allen generierten Inhalt, sofern keine definierte Abbruchbedingung greift.

Dabei gilt:

* generierte ToC-Zeilen sollen entfernt werden
* generierte Config-Zeilen sollen entfernt werden
* zusätzliche, nicht generierte Zeilen im Container sollen erhalten bleiben, aber nicht im Container verbleiben
* verwaltete Heading-Artefakte wie Nummerierung und generierte Inline-Anker sollen entfernt werden

`strip --raw` ist kein toleranterer Modus als `generate`, sondern nutzt denselben Bereinigungsrahmen für beschädigte Managed-Strukturen. Es endet nur in einem anderen Zielzustand.

Die definierten Abbruchbedingungen dieses Bereinigungsrahmens sind in dieser Issue mindestens:

* defekte äußere Container-Struktur
* Excluded bzw. Ignored Regions innerhalb des Containers

## Zu behandelnde Container-Zustände

Die Analyse muss mindestens die folgenden Fälle unterscheiden.

### Äußere Container-Struktur defekt

Beispiele:

* nur Startmarker vorhanden
* nur Endmarker vorhanden
* Startmarker nach Endmarker
* doppelte äußere Marker

Erwartetes Verhalten:

* Abbruch
* Meldung mit Zeilennummern bzw. betroffenen Markerpositionen
* die Fehlermeldung soll die betroffenen Zeilen als Liste im Format `Zeile:Inhalt` ausgeben
* keine schreibende Aktion, auch nicht für `generate` oder `strip --raw`

### Ausgeschlossene oder ignorierte Bereiche innerhalb des Containers

Wenn innerhalb des erkannten Containers ausgeschlossene oder ignorierte Bereiche vorkommen, soll dies nicht stillschweigend akzeptiert werden.

Erwartetes Verhalten:

* Abbruch
* Meldung mit Zeilennummern

Begründung:

* Ein verwalteter Bereich muss strukturell klar und vollständig analysierbar bleiben.
* Ignorierte oder ausgeschlossene Bereiche innerhalb des Containers machen die Zuordnung zwischen generiertem und fremdem Inhalt unnötig unsicher.

### Zusätzliche nicht generierte Zeilen innerhalb des Containers

Wenn innerhalb des Containers zusätzliche nicht generierte Zeilen vorkommen, ist die äußere Container-Struktur zwar nicht zwingend defekt, der Container selbst ist aber nicht intakt.

Erwartetes Verhalten:

* `regen` und `strip` brechen reportend ab
* `check` reportet nur
* `generate` und `strip --raw` sollen die zusätzlichen Zeilen erhalten, aus dem Container herausnehmen und standardmäßig direkt nach dem Container wieder ausgeben

### Äußere Config-Block-Struktur defekt

Beispiele:

* Config-Start ohne Config-Ende
* Config-Ende ohne Config-Start
* doppelter Config-Start
* doppeltes Config-Ende
* falsche Reihenfolge

Erwartetes Verhalten:

* Abbruch mit Zeilennummerninfo
* die Fehlermeldung soll die betroffenen Zeilen als Liste im Format `Zeile:Inhalt` ausgeben
* `regen` und `strip` schreiben nicht
* `check` reportet nur
* `generate` und `strip --raw` dürfen nur dann tolerant fortfahren, wenn keine defekte äußere Container-Struktur vorliegt

### Config-Block außerhalb des Containers

Hier ist die äußere Config-Block-Struktur zwar grundsätzlich erkennbar, der Block liegt aber ganz oder teilweise außerhalb des Containers.

Erwartetes Verhalten:

* die generierten Config-Zeilen dieses Blocks gelten als entfernbar
* nicht generierte Zusatzzeilen sollen nicht pauschal gelöscht werden
* `generate` und `strip --raw` sollen diesen Block aktiv bereinigen
* nicht generierte Zusatzzeilen sollen dabei erhalten bleiben und in gleicher Reihenfolge standardmäßig direkt nach dem Container ausgegeben werden
* `regen`, `strip` und `check` sollen diesen Zustand reporten statt von einem gültigen Managed State auszugehen

### Config-Block-Struktur äußerlich intakt, aber inhaltlich inkonsistent

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
* zusätzliche nicht generierte Zeilen innerhalb dieses Config-Kontexts sollen erhalten bleiben und nach Entfernung des generierten Blocks standardmäßig direkt nach dem Container wieder erscheinen

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
* `regen`: nur bei konsistentem Config-Block und intaktem Container schreiben
* `strip`: nur bei konsistentem Config-Block und intaktem Container schreiben
* `generate`: mit denselben Strukturregeln bereinigen wie `strip --raw` und anschließend neu aufbauen
* `strip --raw`: mit denselben Strukturregeln bereinigen und anschließend den Managed-Zustand vollständig entfernen

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
* Diese gemeinsame Analyse bzw. Integritätsprüfung wird verpflichtend von allen Unterkommandos verwendet.
* Diese Analyse überspringt ausgeschlossene bzw. ignorierte Bereiche im ersten Lauf.
* Die Analyse unterscheidet äußere Strukturfehler, fehlende Container-Intaktheit, Lagefehler und inhaltliche Inkonsistenzen.

### Kommandoverhalten

* `check` schreibt nie.
* `regen` schreibt nur bei konsistentem Config-Block und intaktem Container.
* `strip` schreibt nur bei konsistentem Config-Block und intaktem Container.
* `generate` verwendet für beschädigte Managed-Strukturen denselben Bereinigungsrahmen wie `strip --raw` und erzeugt danach den Zielzustand vollständig neu.
* `strip --raw` entfernt generierte Inhalte auch dann, wenn der Config-Block nicht mehr vollständig parsebar ist, solange keine definierte Abbruchbedingung greift.

### Erhaltung nicht generierter Inhalte

* Zusätzliche Zeilen im Container bleiben erhalten, sofern sie nicht eindeutig generiert sind, machen den Container aber nicht intakt.
* Zusätzliche Zeilen im Config-Block werden nicht pauschal als frei löschbar behandelt, außer soweit sie eindeutig als generierte Config-Zeilen klassifizierbar sind oder die konkrete Regel dies erlaubt.
* Solche zusätzlichen Zeilen bleiben in gleicher Reihenfolge erhalten und sollen nach einer Bereinigung standardmäßig direkt nach dem Container ausgegeben werden.

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
* zusätzliche nicht generierte Zeilen innerhalb eines ansonsten äußerlich intakten Containers

## Umsetzungshinweis

Die Umsetzung erscheint grundsätzlich machbar, wenn sie als Erweiterung der bestehenden Parser- und Fallback-Logik erfolgt und nicht als vollständiger Neuansatz. Besonders naheliegend ist:

* den heutigen strikten Parse-Pfad für gültige Dokumente beizubehalten
* davor oder daneben einen toleranteren Struktur-Scan einzuführen
* die Aktionsentscheidung pro Kommando auf einen gemeinsamen Analyse-Report zu stützen

So kann die bestehende, bereits getestete Logik für den Normalfall weitgehend stabil bleiben, während die Sonderfälle systematischer behandelt werden.

## Weitere Festlegungen für die Umsetzung

Die folgenden Punkte gelten für diese Issue als festgelegt:

* Das Fehlerformat `Zeile:Inhalt` gilt sowohl für defekte äußere Container-Strukturen als auch für defekte äußere Config-Block-Strukturen.
* Falls zusätzlich zu einem defekten Bereich noch ein vollständig gültiger Container existiert, sollen `generate` und `strip --raw` nur den defekten Bereich bereinigen, sofern die äußere Container-Struktur des gültigen Containers selbst nicht defekt ist.
* Nicht generierte Zusatzzeilen sollen nach einer Bereinigung standardmäßig direkt nach dem Container ausgegeben werden; diese Position gilt hier nicht nur als Beispiel, sondern als gewünschte Standardregel.

## Bedeutung für die Spezifikation

Die hier definierte Logik ist nicht nur Implementierungsdetail. Sie sollte nach erfolgreicher Umsetzung auch in die `mdtoc`-Spezifikation überführt werden, insbesondere für:

* Container-Zustände
* Kommandoverhalten bei inkonsistenten Managed-Strukturen
* tolerante Bereinigung durch `strip --raw`
* Verhältnis zwischen strukturellem Scan und semantischer Validierung

---

>## Assistant
>
> Hier ist ein Vorschlag für eine kontrollierte Zerlegung in **mehrere testbare Sub-Issues**, so dass die Gesamtaufgabe nicht in eine unüberschaubare Parser-Refaktorierung kippt.
>
>Ich würde die Arbeit nicht primär nach Kommandos splitten, sondern nach **Analysemodell → Sicherheitsregeln → Bereinigung → Kommandoanbindung → Tests/Spezifikation**. Sonst besteht die Gefahr, dass `check`, `regen`, `strip`, `generate` und `strip --raw` wieder jeweils eigene Sonderlogik bekommen.

---

# Vorgeschlagene Sub-Issues

## Sub-Issue 1: Gemeinsames Analysemodell für `mdtoc`-Container einführen

### Ziel

Ein zentrales Analysemodell einführen, das vor jeder Aktion den Zustand eines Dokuments beziehungsweise eines vorhandenen `mdtoc`-Containers beschreibt.

Diese Issue soll noch **keine Verhaltensänderung der Kommandos** erzwingen, sondern die Grundlage schaffen.

### Umfang

Implementieren einer gemeinsamen Analysefunktion, zum Beispiel konzeptionell:

```text
analyze_managed_container(document) -> ContainerAnalysisReport
```

Der Report sollte mindestens enthalten:

```text
ContainerAnalysisReport
- outer_container_state
- container_marker_positions
- config_outer_state
- config_marker_positions
- config_location_state
- config_semantic_state
- container_integrity_state
- generated_toc_lines
- generated_config_lines
- preserved_extra_lines
- ignored_or_excluded_regions_inside_container
- diagnostics
```

### Zu unterscheidende Zustände

Mindestens:

```text
OuterContainerState
- absent
- intact
- defective
- ambiguous

ContainerIntegrityState
- intact
- not_intact
- not_applicable

OuterConfigState
- absent
- intact
- defective
- ambiguous

ConfigLocationState
- inside_container
- outside_container
- partially_outside_container
- wrong_position
- not_applicable

ConfigSemanticState
- consistent
- inconsistent
- unparseable
- absent
- not_applicable
```

### Akzeptanzkriterien

- Es gibt eine zentrale Analysefunktion.
- Die Analyse ist noch unabhängig vom konkreten Kommando.
- Die Analyse liefert strukturierte Zustände statt nur boolescher Werte.
- Die Analyse enthält Positionsinformationen mit Zeilennummern und Zeileninhalt.
- Bestehende gültige Dokumente werden weiterhin als gültig erkannt.
- Noch keine größeren Schreiboperationen werden geändert.

### Tests

Unit-Tests für den Analyse-Report bei:

- keinem Container
- gültigem Container
- nur Startmarker
- nur Endmarker
- Startmarker nach Endmarker
- doppeltem Startmarker
- doppeltem Endmarker
- gültigem Container mit zusätzlicher fremder Zeile
- Config-Block fehlt
- Config-Block defekt
- Config-Block außerhalb des Containers
- Config-Block teilweise außerhalb des Containers

### Abhängigkeiten

Keine.

---

## Sub-Issue 2: Einheitlicher Struktur-Scan unter Berücksichtigung ignorierter und ausgeschlossener Bereiche

### Ziel

Der erste Analyse-Lauf muss Container- und Config-Strukturen nur außerhalb von ignorierten beziehungsweise ausgeschlossenen Bereichen erkennen.

### Umfang

Der Struktur-Scan soll folgende Bereiche beim ersten Erkennen von Markern aussparen:

- fenced code blocks
- `mdtoc off` / `mdtoc on`-Bereiche
- weitere bereits bestehende ignored/excluded regions der Parser-Logik

Wichtig: Wenn nach Erkennung eines äußeren Containers innerhalb seiner Grenzen ignored/excluded regions liegen, muss dies separat als problematischer Zustand gemeldet werden.

### Regeln

- Marker innerhalb ignored/excluded regions zählen nicht als Container- oder Config-Marker.
- Liegt eine ignored/excluded region innerhalb eines erkannten Containers, darf der Container nicht tolerant bereinigt werden.
- Eine ignored/excluded region innerhalb des Containers ist eine definierte Abbruchbedingung für schreibende Aktionen einschließlich `generate` und `strip --raw`.

### Akzeptanzkriterien

- Container- und Config-Marker in Code-Fences werden ignoriert.
- Container- und Config-Marker in `mdtoc off/on`-Bereichen werden ignoriert.
- Ignored/excluded regions innerhalb eines erkannten Containers werden erkannt und diagnostiziert.
- Der Analyse-Report enthält die betroffenen Zeilenbereiche.
- Keine schreibende Aktion darf bei ignored/excluded regions innerhalb des Containers später fortfahren.

### Tests

Jeweils mit `check`, später zusätzlich mit schreibenden Kommandos:

- `<!-- mdtoc -->` innerhalb eines Code-Fence wird ignoriert.
- `<!-- /mdtoc -->` innerhalb eines Code-Fence wird ignoriert.
- vollständiger Container innerhalb eines `mdtoc off/on`-Blocks wird ignoriert.
- Code-Fence innerhalb eines echten Containers führt zu Diagnose.
- `mdtoc off/on`-Bereich innerhalb eines echten Containers führt zu Diagnose.

### Abhängigkeiten

Sub-Issue 1.

---

## Sub-Issue 3: Erkennung eindeutig generierter ToC-Zeilen implementieren oder zentralisieren

### Ziel

ToC-Zeilen sollen zuverlässig als generiert oder nicht generiert klassifiziert werden, ohne den aktuellen Dokumentinhalt als Maßstab zu verwenden.

### Umfang

Eine zentrale Klassifikationsfunktion einführen, zum Beispiel:

```text
classify_toc_line(line, slug_mode, duplicate_heading_context) -> generated | non_generated | blank
```

### Regeln

Eine ToC-Zeile gilt als eindeutig generiert, wenn:

- führende und abschließende Whitespaces toleriert werden
- ein Bullet vorhanden ist
- danach eine Markdown-Link-Struktur `[text](slug)` folgt
- der Link-Slug plausibel zum Text passt
- doppelte Überschriften beziehungsweise Slug-Suffixe berücksichtigt werden
- bei nicht parsebarem Config-Block standardmäßig GitHub-Slug-Verhalten angenommen wird

Nicht erforderlich:

- Die ToC-Zeile muss nicht auf eine aktuell vorhandene Überschrift im Dokument zeigen.
- Der ToC muss nicht aktuell sein.

### Akzeptanzkriterien

- Veraltete, aber syntaktisch gültige ToC-Zeilen werden als generiert erkannt.
- ToC-Zeilen mit trailing whitespace werden als generiert erkannt.
- Beliebige Nutzerzeilen im ToC-Bereich werden nicht als generiert klassifiziert.
- Falsch gesluggte Links werden nicht als sicher generiert klassifiziert.
- Die Klassifikation ist unabhängig vom konkreten Kommando.

### Tests

- `* [Intro](#intro)` → generiert
- `  * [2.1. Releases](#releases)` → generiert
- `  * [API API](#api-api)` → generiert
- `  * [API](#api-1)` → generiert, wenn Duplicate-Slug-Kontext plausibel ist
- `* [Intro](#wrong)` → nicht generiert
- `hello world` im ToC-Bereich → nicht generiert
- Leerzeile → blank
- Whitespace-only-Zeile → blank

### Abhängigkeiten

Sub-Issue 1.

---

## Sub-Issue 4: Erkennung generierter und nicht generierter Config-Kontext-Zeilen

### Ziel

Config-Zeilen innerhalb eines Config-Kontexts sollen strukturell klassifizierbar sein, auch wenn der Config-Block semantisch nicht parsebar oder inkonsistent ist.

### Umfang

Zentrale Klassifikation:

```text
classify_config_context_line(line) -> generated_config_line | non_generated | blank | marker
```

### Regeln

Generierte Config-Zeilen sind:

```text
key=value
```

mit optionalem trailing whitespace.

Nicht generierte Zusatzzeilen im Config-Kontext dürfen nicht pauschal gelöscht werden.

### Wichtig

Diese Issue unterscheidet nur strukturell:

- Ist die Zeile grundsätzlich eine Config-Zeile?
- Ist sie eindeutig nicht generierter Inhalt?
- Ist sie leer?

Die semantische Validierung, ob Schlüssel erwartet, Reihenfolge korrekt und Werte gültig sind, erfolgt separat.

### Akzeptanzkriterien

- `key=value` wird als Config-Zeile erkannt.
- `key=value   ` wird als Config-Zeile erkannt.
- `this is a user note` wird als nicht generiert klassifiziert.
- Leere Zeilen werden als blank klassifiziert.
- Config-Marker werden separat klassifiziert.
- Nicht generierte Zeilen im Config-Kontext werden im Analyse-Report als zu erhaltende Zusatzzeilen ausgewiesen.

### Tests

- gültige Config-Zeilen
- unbekannte, aber syntaktisch generierte Config-Zeilen
- beschädigte Config-Zeilen
- Nutzertext im Config-Block
- Leerzeilen
- Config-Start und Config-Ende

### Abhängigkeiten

Sub-Issue 1.

---

## Sub-Issue 5: Semantische Config-Validierung vom Struktur-Scan trennen

### Ziel

Der bestehende strikte Config-Parse-Pfad soll weiterverwendet werden, aber erst nach dem toleranten Struktur-Scan greifen.

### Umfang

Die semantische Validierung prüft nur, wenn die äußere Config-Struktur formal verwertbar ist.

Zu prüfen:

- Config-Block liegt vollständig im Container
- Config-Block steht an der vorgesehenen Position
- erwartete Schlüssel vorhanden
- Reihenfolge korrekt
- relevante Zeilenanzahl korrekt
- Werte gültig

### Ergebniszustände

```text
ConfigSemanticState
- consistent
- inconsistent
- unparseable
- absent
- not_applicable
```

### Akzeptanzkriterien

- Ein formal intakter und gültiger Config-Block wird als `consistent` erkannt.
- Falsche Schlüssel werden als `inconsistent` erkannt.
- Falsche Reihenfolge wird als `inconsistent` erkannt.
- Ungültige Werte werden als `inconsistent` erkannt.
- Defekte äußere Config-Struktur wird nicht semantisch geparst, sondern als strukturelles Problem gemeldet.
- Der bestehende Normalfall bleibt stabil.

### Tests

- gültige Config
- fehlender Schlüssel
- zusätzlicher generierter Schlüssel
- falsche Reihenfolge
- ungültiger Wert
- falsche Zeilenanzahl
- Config-Block außerhalb des Containers
- Config-Block an falscher Position

### Abhängigkeiten

Sub-Issue 1, 4.

---

## Sub-Issue 6: Einheitliche Diagnose- und Fehlerausgabe für Strukturprobleme

### Ziel

Strukturfehler sollen konsistent, verständlich und maschinen-/testbar ausgegeben werden.

### Umfang

Einheitliches Diagnosemodell, zum Beispiel:

```text
Diagnostic
- severity
- code
- message
- affected_lines
- command_action
```

Für Markerfehler muss die Ausgabe betroffene Zeilen im Format enthalten:

```text
Zeile:Inhalt
```

Beispiel:

```text
Mehrdeutige mdtoc-Containermarker; bitte manuell korrigieren oder entfernen:
1: <!-- mdtoc -->
17: <!-- /mdtoc -->
42: <!-- /mdtoc -->
```

### Diagnoseklassen

Mindestens:

```text
defective_outer_container
ambiguous_outer_container
ignored_region_inside_container
defective_outer_config
config_outside_container
config_wrong_position
config_inconsistent
container_not_intact
extra_non_generated_lines
```

### Akzeptanzkriterien

- Defekte äußere Container-Strukturen enthalten Zeilennummern und Zeileninhalt.
- Defekte äußere Config-Strukturen enthalten Zeilennummern und Zeileninhalt.
- Lageprobleme des Config-Blocks werden als solche benannt.
- Inhaltliche Inkonsistenzen werden von Strukturproblemen unterscheidbar gemeldet.
- Tests können auf Diagnose-Codes prüfen, nicht nur auf Freitext.

### Tests

- nur Startmarker
- nur Endmarker
- doppelte Marker
- Config-Start ohne Config-Ende
- Config-Ende ohne Config-Start
- doppelter Config-Start
- doppeltes Config-Ende
- Config außerhalb Container
- Zusatzzeilen im Container

### Abhängigkeiten

Sub-Issue 1, 5.

---

## Sub-Issue 7: Aktionsmatrix für alle Kommandos zentralisieren

### Ziel

Alle Kommandos sollen aus demselben Analyse-Report ihre Aktionsrechte ableiten.

### Umfang

Ein zentrales Entscheidungselement einführen, zum Beispiel:

```text
decide_action(command, analysis_report) -> ActionDecision
```

### Matrix

| Zustand | `check` | `regen` | `strip` | `generate` | `strip --raw` |
|---|---:|---:|---:|---:|---:|
| Äußerer Container defekt | report | abort | abort | abort | abort |
| Ignored/excluded region im Container | report | abort | abort | abort | abort |
| Container intakt, Config konsistent | report/compare | write allowed | write allowed | write allowed | write allowed |
| Container nicht intakt wegen Zusatzzeilen | report | abort | abort | tolerant clean + write | tolerant clean |
| Config äußerlich defekt | report | abort | abort sofern Containerstruktur defekt; sonst tolerant möglich | tolerant möglich |
| Config außerhalb Container | report | abort | abort | tolerant clean + write | tolerant clean |
| Config semantisch inkonsistent | report | abort | abort | tolerant clean + write | tolerant clean |
| Kein Container vorhanden | bisheriges Verhalten | bisheriges Verhalten | bisheriges Verhalten | normal generate | no-op/aktuelles Verhalten prüfen |

### Akzeptanzkriterien

- `check`, `regen`, `strip`, `generate`, `strip --raw` verwenden dieselbe Entscheidungsfunktion.
- `regen` schreibt nur bei intaktem Container und konsistentem Config-Block.
- `strip` schreibt nur bei intaktem Container und konsistentem Config-Block.
- `generate` und `strip --raw` verwenden dieselben strukturellen Abbruchbedingungen.
- Die Entscheidung ist testbar, ohne echte Dateien schreiben zu müssen.

### Tests

Reine Unit-Tests für die Matrix:

- pro Analysezustand
- pro Kommando
- erwartete Entscheidung: `report_only`, `abort`, `write_allowed`, `tolerant_clean_then_write`, `tolerant_clean_then_strip`

### Abhängigkeiten

Sub-Issue 1, 6.

---

## Sub-Issue 8: Tolerante Bereinigung für `generate` und `strip --raw` implementieren

### Ziel

Beschädigte, aber strukturell ausreichend erkennbare Managed-Strukturen sollen sicher bereinigt werden, ohne nicht generierten Inhalt zu verlieren.

### Umfang

Eine gemeinsame Bereinigungsfunktion einführen, zum Beispiel:

```text
clean_damaged_managed_state(document, analysis_report) -> CleanedDocument
```

Diese Funktion wird von `generate` und `strip --raw` gemeinsam verwendet.

### Regeln

Entfernt werden dürfen:

- eindeutig generierte ToC-Zeilen
- eindeutig generierte Config-Zeilen
- Container- und Config-Marker, sofern die Struktur ausreichend erkannt wurde
- Leerzeilen nach definierter Normalisierung
- verwaltete Heading-Artefakte bei `strip --raw`

Erhalten werden müssen:

- nicht generierte Zusatzzeilen im ToC-Bereich
- nicht generierte Zusatzzeilen im Config-Kontext
- nicht generierte Zusatzzeilen innerhalb eines äußerlich intakten Containers
- Reihenfolge dieser Zeilen

Standardplatzierung:

```text
direkt nach dem neu aufgebauten beziehungsweise entfernten Containerbereich
```

Für `generate` bedeutet das:

```text
[neuer normalisierter Container]
[erhaltene nicht generierte Zusatzzeilen]
```

Für `strip --raw` bedeutet das:

```text
[erhaltene nicht generierte Zusatzzeilen]
```

an der Stelle des entfernten Containers beziehungsweise direkt nach der ehemaligen Containerposition.

### Akzeptanzkriterien

- Nicht generierte Zusatzzeilen werden nicht gelöscht.
- Nicht generierte Zusatzzeilen verbleiben nicht im neuen Container.
- Reihenfolge der Zusatzzeilen bleibt erhalten.
- Generierte ToC-Zeilen werden entfernt.
- Generierte Config-Zeilen werden entfernt.
- Leerzeilen werden konsistent normalisiert.
- `generate` und `strip --raw` teilen denselben Bereinigungsrahmen.
- Bei ignored/excluded regions innerhalb des Containers wird nicht bereinigt, sondern abgebrochen.

### Tests

Für `generate` und `strip --raw`:

- Zusatzzeile im ToC-Bereich
- mehrere Zusatzzeilen im ToC-Bereich
- Zusatzzeile im Config-Block
- Zusatzzeilen im ToC- und Config-Bereich
- kaputte Config mit Nutzerzeile
- Config außerhalb Container mit Nutzerzeile
- generierte ToC- und Config-Reste werden entfernt
- Reihenfolge mehrerer Zusatzzeilen bleibt erhalten
- Zusatzzeilen landen direkt nach dem Container beziehungsweise an definierter Ersatzposition

### Abhängigkeiten

Sub-Issue 3, 4, 7.

---

## Sub-Issue 9: `check` auf gemeinsame Analyse umstellen

### Ziel

`check` soll ausschließlich den gemeinsamen Analyse-Report verwenden und niemals schreiben.

### Umfang

`check` soll:

- Analyse durchführen
- Inkonsistenzen reporten
- Fehlerstatus setzen, falls erforderlich
- keine Inhalte verschieben, löschen oder normalisieren

### Akzeptanzkriterien

- `check` verändert keine Datei.
- `check` meldet defekte äußere Container-Strukturen.
- `check` meldet defekte Config-Strukturen.
- `check` meldet Config außerhalb Container.
- `check` meldet nicht intakten Container wegen Zusatzzeilen.
- `check` unterscheidet Strukturfehler, Lagefehler und semantische Inkonsistenz.

### Tests

Für alle relevanten Zustände:

- Dateiinhalt vor/nach `check` identisch
- erwarteter Exit-Code
- erwartete Diagnose-Codes
- erwartete Zeilenausgabe bei Markerproblemen

### Abhängigkeiten

Sub-Issue 1, 6, 7.

---

## Sub-Issue 10: `regen` und normales `strip` auf strikte Aktionsrechte umstellen

### Ziel

`regen` und normales `strip` dürfen nur bei intaktem Container und konsistentem Config-Block schreiben.

### Umfang

Beide Kommandos verwenden:

```text
analyze_managed_container()
decide_action()
```

Schreiben ist nur erlaubt, wenn:

```text
container_integrity_state == intact
config_semantic_state == consistent
```

### Akzeptanzkriterien

- `regen` schreibt bei gültigem Container wie bisher.
- `strip` schreibt bei gültigem Container wie bisher.
- `regen` bricht bei Zusatzzeilen im Container ab.
- `strip` bricht bei Zusatzzeilen im Container ab.
- Beide brechen bei defekter Config ab.
- Beide brechen bei Config außerhalb Container ab.
- Beide brechen bei semantisch inkonsistenter Config ab.
- Beide brechen bei ignored/excluded regions innerhalb Container ab.
- Kein stilles Bereinigen in diesen beiden Kommandos.

### Tests

Jeweils für `regen` und `strip`:

- gültiger Container → schreibt
- Zusatzzeile im ToC-Bereich → abort
- Zusatzzeile im Config-Kontext → abort
- defekter Config-Start → abort
- Config außerhalb Container → abort
- ungültiger Config-Wert → abort
- ignored region im Container → abort
- äußerer Container defekt → abort

### Abhängigkeiten

Sub-Issue 7, 9.

---

## Sub-Issue 11: `generate` auf toleranten Bereinigungsrahmen umstellen

### Ziel

`generate` soll vorhandene beschädigte Managed-Strukturen sicher bereinigen und danach einen neuen vollständigen Container rendern.

### Umfang

`generate` verwendet:

```text
analyze_managed_container()
decide_action()
clean_damaged_managed_state()
render_new_container()
```

### Regeln

- Wenn Config konsistent ist, werden vorhandene Config-Werte mit CLI-Werten gemergt.
- CLI-Werte haben Vorrang.
- Wenn Config nicht konsistent oder nicht verwertbar ist, wird mit Default-/CLI-Werten neu generiert.
- Bei beschädigtem, aber äußerlich verwertbarem Container wird bereinigt.
- Nicht generierte Zusatzzeilen werden nach dem neuen Container ausgegeben.
- Bei defekter äußerer Container-Struktur wird abgebrochen.
- Bei ignored/excluded regions innerhalb Container wird abgebrochen.

### Akzeptanzkriterien

- `generate` erstellt bei gültigem Dokument weiterhin den erwarteten Container.
- `generate` ersetzt beschädigte Managed-Strukturen durch einen neuen normalisierten Container.
- Nicht generierte Zusatzzeilen bleiben erhalten.
- Zusatzzeilen landen direkt nach dem neu aufgebauten Container.
- Generierte ToC- und Config-Reste verschwinden.
- CLI-Werte überschreiben verwertbare Config-Werte.
- Bei nicht verwertbarer Config wird nicht versucht, kaputte Werte heuristisch zu retten.
- Bei Strukturabbruchbedingungen wird nicht geschrieben.

### Tests

- kein Container → normal generate
- gültiger Container → normal regenerate
- zusätzlicher Nutzertext im ToC-Bereich → neuer Container + Nutzertext danach
- Nutzertext im Config-Block → neuer Container + Nutzertext danach
- Config inkonsistent → neuer Container aus CLI/defaults
- Config außerhalb Container → bereinigen + neuer Container
- defekter äußerer Container → abort
- ignored region im Container → abort

### Abhängigkeiten

Sub-Issue 8, 10.

---

## Sub-Issue 12: `strip --raw` auf denselben Bereinigungsrahmen wie `generate` umstellen

### Ziel

`strip --raw` soll denselben beschädigungstoleranten Strukturrahmen wie `generate` nutzen, aber keinen neuen Managed-Zustand aufbauen.

### Umfang

`strip --raw` verwendet:

```text
analyze_managed_container()
decide_action()
clean_damaged_managed_state()
remove_heading_artifacts()
```

### Regeln

- Generierte ToC-Zeilen entfernen.
- Generierte Config-Zeilen entfernen.
- Container-/Config-Marker entfernen, sofern sicher erkannt.
- Nicht generierte Zusatzzeilen erhalten.
- Zusatzzeilen an definierter Position ausgeben.
- Verwaltete Heading-Artefakte entfernen.
- Bei defekter äußerer Container-Struktur abbrechen.
- Bei ignored/excluded regions innerhalb Container abbrechen.

### Akzeptanzkriterien

- `strip --raw` entfernt Managed-Strukturen auch bei inkonsistenter Config.
- `strip --raw` entfernt Managed-Strukturen auch bei Config außerhalb Container, sofern keine harte Abbruchbedingung greift.
- Nicht generierte Zusatzzeilen bleiben erhalten.
- Zusatzzeilen werden nicht im Container belassen, weil der Container entfernt wird.
- Heading-Artefakte werden wie bisher beziehungsweise gemäß Raw-Regeln entfernt.
- `strip --raw` und `generate` teilen dieselben Abbruchbedingungen.

### Tests

- gültiger Container → vollständig entfernt
- inkonsistente Config → generierte Inhalte entfernt
- Config außerhalb Container → generierte Config entfernt
- Zusatzzeile im ToC-Bereich → bleibt erhalten
- Zusatzzeile im Config-Kontext → bleibt erhalten
- defekter äußerer Container → abort
- ignored region im Container → abort
- Heading-Nummerierung wird entfernt
- generierte Inline-Anker werden entfernt

### Abhängigkeiten

Sub-Issue 8, 11.

---

## Sub-Issue 13: Behandlung von Config-Block außerhalb des Containers

### Ziel

Dieser Sonderfall ist wichtig genug für eine eigene Issue, weil er leicht zu Datenverlust führen kann.

### Umfang

Wenn eine äußerlich intakte Config-Struktur ganz oder teilweise außerhalb des Containers liegt:

- generierte Config-Zeilen dürfen entfernt werden
- nicht generierte Zusatzzeilen im Config-Kontext müssen erhalten bleiben
- `check`, `regen`, `strip` reporten beziehungsweise brechen ab
- `generate` und `strip --raw` bereinigen tolerant, sofern keine harte Abbruchbedingung greift

### Akzeptanzkriterien

- Config vor Container wird erkannt.
- Config nach Container wird erkannt.
- Config beginnt im Container und endet außerhalb → Lagefehler.
- Config beginnt außerhalb und endet im Container → Lagefehler.
- Nicht generierte Zeilen im externen Config-Kontext bleiben erhalten.
- Generierte Config-Zeilen werden bei `generate` und `strip --raw` entfernt.
- `regen` und normales `strip` schreiben nicht.

### Tests

Für alle Kommandos:

- Config vollständig vor Container
- Config vollständig nach Container
- Config teilweise außerhalb
- externe Config mit Nutzerzeile
- externe Config nur mit generierten Config-Zeilen
- externe Config mit Leerzeilen

### Abhängigkeiten

Sub-Issue 5, 8.

---

## Sub-Issue 14: Harte Abbruchbedingungen vollständig absichern

### Ziel

Bestimmte Zustände dürfen niemals tolerant bereinigt werden, weil die Gefahr von Datenverlust oder Fehlzuordnung zu groß ist.

### Harte Abbruchbedingungen

Mindestens:

- defekte äußere Container-Struktur
- mehrdeutige äußere Container-Marker
- ignored/excluded region innerhalb des erkannten Containers

Optional zusätzlich explizit prüfen:

- überlappende Container
- verschachtelte Container
- nicht eindeutig zuordenbare Container-/Config-Marker-Kombinationen

### Akzeptanzkriterien

- `generate` bricht bei harter Abbruchbedingung ab.
- `strip --raw` bricht bei harter Abbruchbedingung ab.
- `regen` und `strip` brechen ebenfalls ab.
- `check` reportet ohne Änderung.
- Fehlermeldung enthält relevante Zeilen.
- Keine Datei wird bei diesen Zuständen geschrieben.

### Tests

Für alle Kommandos:

- nur Startmarker
- nur Endmarker
- doppelter Startmarker
- doppelter Endmarker
- Startmarker nach Endmarker
- Code-Fence innerhalb Container
- `mdtoc off/on` innerhalb Container
- verschachtelter Container, falls relevant
- überlappende Config-/Container-Strukturen

### Abhängigkeiten

Sub-Issue 2, 6, 7.

---

## Sub-Issue 15: Vollständige End-to-End-Testmatrix für alle Kommandos

### Ziel

Nach den Unit-Tests der Einzelkomponenten braucht es eine überschaubare, aber vollständige E2E-Matrix.

### Umfang

Für jeden relevanten Zustand mindestens Tests für:

- `check`
- `regen`
- `strip`
- `generate`
- `strip --raw`

### Zustände

Mindestens:

1. gültiger Container
2. äußerer Container defekt
3. ignored/excluded region im Container
4. zusätzliche nicht generierte Zeile im ToC-Bereich
5. zusätzliche nicht generierte Zeile im Config-Kontext
6. Config-Block defekt
7. Config-Block außerhalb Container
8. Config-Block semantisch inkonsistent
9. kein Container
10. generierte ToC-Reste ohne gültige Config, sofern vom aktuellen Verhalten relevant

### Akzeptanzkriterien

- Jede Kombination aus Kommando und Zustand hat einen expliziten erwarteten Ausgang.
- Schreibende Kommandos werden auf Dateiinhalt geprüft.
- Reportende Kommandos werden auf Exit-Code und Diagnose geprüft.
- Keine Tests verlassen sich nur auf Freitext, wo Diagnose-Codes verfügbar sind.
- Datenverlust-Szenarien werden explizit getestet.

### Abhängigkeiten

Sub-Issue 9–14.

---

## Sub-Issue 16: Spezifikation aktualisieren

### Ziel

Die neue Logik ist Teil des Formats und der Nutzererwartung, nicht nur Implementierungsdetail.

### Umfang

Dokumentieren:

- Container-Zustände
- äußere Container-Struktur vs. Container-Intaktheit
- äußere Config-Struktur vs. Config-Konsistenz
- ignorierte/excluded Bereiche
- harte Abbruchbedingungen
- Verhalten je Kommando
- tolerante Bereinigung durch `generate` und `strip --raw`
- Erhaltung nicht generierter Zusatzzeilen
- Positionierung erhaltener Zusatzzeilen direkt nach dem Container
- Fehlerformat `Zeile:Inhalt`

### Akzeptanzkriterien

- Spezifikation beschreibt dieselben Zustände wie der Analyse-Report.
- Kommandoverhalten ist tabellarisch oder eindeutig beschrieben.
- Datenverlustschutz ist explizit dokumentiert.
- `generate` und `strip --raw` werden als gleicher Bereinigungsrahmen mit unterschiedlichem Zielzustand beschrieben.
- Beispiele für mindestens drei beschädigte Containerfälle sind enthalten.

### Abhängigkeiten

Sub-Issue 7–15.

---

# Empfohlene Reihenfolge

Ich würde die Umsetzung in dieser Reihenfolge planen:

1. **Sub-Issue 1**: Analysemodell  
2. **Sub-Issue 2**: ignored/excluded handling  
3. **Sub-Issue 3**: ToC-Zeilenklassifikation  
4. **Sub-Issue 4**: Config-Zeilenklassifikation  
5. **Sub-Issue 5**: Semantische Config-Validierung  
6. **Sub-Issue 6**: Diagnosemodell  
7. **Sub-Issue 7**: Aktionsmatrix  
8. **Sub-Issue 9**: `check` anbinden  
9. **Sub-Issue 10**: `regen` und `strip` anbinden  
10. **Sub-Issue 8**: Gemeinsame tolerante Bereinigung  
11. **Sub-Issue 11**: `generate` anbinden  
12. **Sub-Issue 12**: `strip --raw` anbinden  
13. **Sub-Issue 13**: Config außerhalb Container härten  
14. **Sub-Issue 14**: Harte Abbruchbedingungen absichern  
15. **Sub-Issue 15**: E2E-Testmatrix  
16. **Sub-Issue 16**: Spezifikation aktualisieren

---

# Alternative kompaktere Aufteilung

Falls ihr weniger Issues wollt, würde ich sie so bündeln:

## A. Gemeinsame Containeranalyse und Zustandsmodell

Enthält Sub-Issues 1–6.

## B. Zentrale Aktionsmatrix und Kommandoanbindung für `check`, `regen`, `strip`

Enthält Sub-Issues 7, 9, 10.

## C. Gemeinsamer Bereinigungsrahmen für `generate` und `strip --raw`

Enthält Sub-Issues 8, 11, 12, 13.

## D. Harte Abbruchbedingungen und E2E-Testmatrix

Enthält Sub-Issues 14, 15.

## E. Spezifikation aktualisieren

Entspricht Sub-Issue 16.

Für kontrollierte Umsetzung und Reviewbarkeit ist die feinere Variante aber besser.

---

# Wichtigste Designentscheidung

Die zentrale Leitlinie sollte in jeder Sub-Issue wieder auftauchen:

> Die Analyse entscheidet über den Zustand. Die Kommandos entscheiden nur noch anhand dieses Zustands, ob sie reporten, abbrechen, strikt schreiben oder tolerant bereinigen dürfen.

Damit wird verhindert, dass `generate`, `strip --raw`, `regen`, `strip` und `check` wieder divergierende Parser- und Reparaturpfade bekommen.

---

