# Issue #54 muss noch einmal neu überdacht werden

Begriffe im Rahmen dieser Issue:

- excluded Regions sind CodeBlocks plus excluded Regions
- Container: ist immer an eine intakte outer Container Structure gebunden
- intakte outer Container Structure ist schon besprochen
- Config Block: ist immer an eine intakte outer Config Block Structure gebunden und muss im Container liegen
- intakte outer Config Block Structure: ahnlich intakte outer Container Structure zu verstehen im Config Block Kontext
- Container besteht aus ToC Bereich plus Config Block plus ggf. zuätzlichen Zeilen, die aus Versehen dort gelandet sind
- ToC Bereich: besteht aus eideutig generierten ToC Zeilen plus ggf. zuätzlichen Zeilen, die aus Versehen dort gelandet sind
- eindeutig generierte ToC Zeilen sehen z.B. so aus: "  * [2.1. Releases](#releases)". Sie starten und enden optional mit Whitespace, gefolgt von einem Bullet mit Leerzeichen, Dann []() Strutur, wobei der Text im () dem aus [] slug generiertem Text entsprechen muss. Liegt kein intakter Config Block vor, wird gihub als Regel angenommemn. Whitespaces am Zeilenende werden zugelassen, da sie im Editor oft nicht sichtbar sind und ggf. Verwirrung hervorrufen, wenn  solche Zeilen als nicht generierte zeilen behandelt werden.
- Config Block besteht aus Config-Zeilen plus ggf. zuätzlichen Zeilen, die aus Versehen dort gelandet sind
- Config Zeilen sind reine key=value Zeilen plus optionale Whitspaces am ende

Generelles Verhalten:

- Kommando `check` ändert in keinem Fall etwas, es reportet nur.
- Kommando `regen` & `strip` dürfen nur agieren, wenn Config Block konsistent ist. Sonst zwangsweise reporten.
- Kommando `generate`, versucht einen intakten Config Block zu lesen umd mit CLI Werten zu mergen, führt dann intern (gedanklich) `strip --raw` aus, um dadach alles wieder neu aufzubauen. Wenn das jetzige Verhalten anders ist, kann es so bleiben, denn es ist getesteter Code. Aber hier darlegen, wie es läuft.
- Kommando `strip --raw` entfernt allen generierten Inhalt, wenn keine Abbruchbedingung vorliegt und lässt zusätzliche Zeilen im Container überleben.

Logik:

- Alle Kommandos müssen zunächst den Container-Zustand ermitteln.
- Alle Container & Config Block Analysen müssen im ersten Run Excluded Regions übersringen
- Der Container kann in verschiedenen Zuständen sein, die generell jeweils behandelt werden müssen:
  - outer structure defect (wie besprochen) -> Abbruch mit Zeilennummerninfo wie besprochen
  - wenn im Container exluded Regions vorkomme -> Abbruch mit Zeilennummern
  - config block outer structure defect (wie bei container outer structure besprochen) -> Abbruch mit Zeilennummerninfo wie bei container outer structure
  - config block outside container -> remove this config block, but only generated config block lines 
  - config block outer structure ok
    - config block inconsistent (Zeilenzahl und/oder Zeileninhalt stimmen nicht)
      - remove generated config block lines

Aufgabe: Analysiere Issue #54 plus den Kommentar zusammen mit diesen Überlegungen hier. Prüfe ob die Herangehensweise umfassend und konsistent ist und mit erträglichem Aufwand umsetzbar. Stelle Rückfragen und mache ggf. Vorschläge. Wenn grundsätzlich soweit alles klar ist, formuliere die Issue komplett neu und ausführlich  im ./docs Folder als separate Markdown-Datei. Wenn diese dann abschließend bearbeitet ist von mir und/oder dir, werde ich deren Inhalt anstelle des jetzigen Issue #54 Inhaltes einsetzen und den jetzigen Kommentar löschen, es dürfen also keine Referenzen darauf gemacht werden. Schreibe zunächt die Datei in Deutsch mit dem Namen Issue54_(german).md. Übersetzung ins Englich ist ein nachfolgender Step. Auch wenn es ein Issue Text ist, soll es sehr ausführlich sein, aber unnötige Widerholungen vermeiden, Widerhlungen nur für die Klarhet. Der Englische Text soll mit Umsetzung der Issue #54 dann auch in die mdtoc-spec einfließen.