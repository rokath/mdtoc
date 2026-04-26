# Issue #54: Robustere Analyse und Behandlung fehlerhafter `mdtoc`-Container

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
* `strip --raw`: mit denselben Strukturregeln bereinigen wie `generate` und anschließend den Managed-Zustand vollständig entfernen

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
