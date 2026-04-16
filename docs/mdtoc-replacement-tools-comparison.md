# 🧭 Überblick: relevante Tools

Tools, die sehr ähnlich sind – mit Fokus auf:

- CLI / CI-tauglich
- multiplattform
- deterministisches Verhalten
- Markdown → ToC / Anchor-Logik

## Vergleichstabelle 1

| Tool                                       | Sprache | CLI / CI     | Anchor-Strategie                         | Besonderheiten              | Link                                                                                            |
|--------------------------------------------|---------|--------------|------------------------------------------|-----------------------------|-------------------------------------------------------------------------------------------------|
| **markdown-toc**                           | Node.js | ✅            | configurable (`slugify`)                 | sehr verbreitet, simpel     | [GitHub Repo öffnen](https://github.com/jonschlinkert/markdown-toc?utm_source=chatgpt.com)      |
| **md-toc**                                 | Python  | ✅            | renderer-kompatibel (GitHub/GitLab etc.) | bewusst standardkonform     | [GitHub Repo öffnen](https://github.com/frnmst/md-toc?utm_source=chatgpt.com)                   |
| **bitdowntoc**                             | Kotlin  | ✅            | Profile (GitHub, GitLab, etc.)           | idempotent + marker-basiert | [GitHub Repo öffnen](https://github.com/derlin/bitdowntoc?utm_source=chatgpt.com)               |
| **doctoc (Python Variante)**               | Python  | ✅            | klassische Slug-Links                    | einfach, CLI                | [GitHub Repo öffnen](https://github.com/ktechhub/doctoc?utm_source=chatgpt.com)                 |
| **Generic Markdown TOC Generator (Tools)** | diverse | ⚠️ meist Web | GitHub-kompatible Slugs                  | eher Referenzverhalten      | [Beispiel Tool ansehen](https://toolsbox.io/code/markdown-toc-generator?utm_source=chatgpt.com) |
| **MkDocs (indirekt)**                      | Python  | ✅            | saubere Slugs                            | kompletter Pipeline-Ansatz  | [MkDocs Infos](https://en.wikipedia.org/wiki/MkDocs?utm_source=chatgpt.com)                     |

> Was diese Tools gemeinsam machen
> 
> Alle relevanten Tools:
> 
> parsen #-Headings
> bauen hierarchischen ToC
> erzeugen Anchors deterministisch
> ignorieren Codeblöcke
> sind CI-fähig
> 
> 👉 Genau dein Problemraum.
> 
> 🔍 Wichtige Unterschiede (entscheidend für dein Design)
> 1. Anchor-Strategie
> dumeng (dein aktueller Zustand)
> fast keine Transformation
> nicht URL-safe
> inkonsistent
> moderne Tools (Standard)
> 
> 👉 fast alle machen Slug-Generierung
> 
> Typisch:
> 
> Übergrößenträger → ubergrossentrager
> naïve façade → naive-facade
> 
> ➡️ URL-safe
> ➡️ stabil
> ➡️ renderer-kompatibel
> 
> 2. Renderer-Kompatibilität
> 
> Einige Tools (z. B. md-toc) gehen noch weiter:
> 
> 👉 passen Verhalten an Plattform an:
> 
> GitHub
> GitLab
> Bitbucket
> 
> 3. Idempotenz (sehr wichtig)
> 
> Bestes Beispiel:
> 
> 👉 bitdowntoc
> 
> erkennt vorhandenen ToC
> überschreibt ihn sauber
> kann beliebig oft laufen
> 
> 👉 exakt dein Ziel.
> 
> 🧠 Was sich für dein Projekt ändert (ohne dumeng-Zwang)
> 
> Jetzt wird es interessant:
> 
> Option 1 — „State of the Art“ (empfohlen)
> 
> 👉 Du gehst auf Slug-basierte Anchors
> 
> Vorteile
> kompatibel mit:
> GitHub
> GitLab
> MkDocs
> stabile Links
> keine kaputten Links wie jetzt
> Option 2 — „Plattform-Modus“
> 
> 👉 wie md-toc:
> 
> mdtoc generate --anchor-style=github
> mdtoc generate --anchor-style=gitlab
> Option 3 — „Minimal (dein Ansatz)“
> 
> 👉 einfacher Parser + dumeng-like
> 
> → aber du wolltest ja weg davon
> 
> 🚀 Meine klare Empfehlung (für dich)
> 
> 👉 Geh auf:
> 
> ✅ Default: slug-basierte Anchors
> 
> Warum:
> 
> Standard in fast allen Tools
> einfacher als dumeng (!) langfristig
> besser testbar
> keine VSCode-Probleme
> 📐 Konkrete Ziel-Definition für mdtoc
> 
> Du könntest definieren:
> 
> Default anchor-style = slug
> 
> Optional:
> 
> --anchor-style=dumeng
> --anchor-style=github
> 🧩 Wie dein Tool sich positioniert
> 
> Nach deiner aktuellen Entwicklung wäre mdtoc:
> 
> 👉 Mischung aus:
> 
> bitdowntoc (Idempotenz + Marker)
> markdown-toc (CLI simplicity)
> md-toc (Renderer-Bewusstsein)
> 
> Das ist eine sehr gute Position.
> 
> ⚠️ Wichtiger Punkt (Designentscheidung!)
> 
> Ich würde diese Frage ins Spec schreiben:
> 
> > Designentscheidung:
> > mdtoc verwendet standardmäßig slug-basierte Anchor-IDs,
> > da diese stabil, URL-sicher und mit gängigen Markdown-Renderern kompatibel sind.
> 🧾 Kurzfazit
> 
> Du bist jetzt an einem Wendepunkt:
> 
> Variante	Bewertung
> dumeng-kompatibel	einfach, aber fehleranfällig
> slug-basiert	Best Practice
> renderer-spezifisch	maximal kompatibel, komplexer
> 
> 👉 Ich würde klar sagen:
> 
> Wechsle zu slug – das ist das professionelle Setup.

# Open-Source-Alternativen zu `mdtoc` für das Trice-Projekt

## Geltungsbereich

Diese Übersicht enthält **nur frei nutzbare Tools**, die sich für OSS-Workflows eignen und sich grundsätzlich in **CLI-/CI-Pipelines** einsetzen lassen.  
**UI** habe ich nicht als Tabellenmerkmal geführt, weil alle gelisteten Kandidaten **CLI-first** sind und **keine klassische Desktop-GUI** mitbringen.  
**dumeng-toc-Kompatibilität** war **kein** Kriterium mehr.

## Wichtige Vorbemerkungen

- **Ein `slug` ist nicht automatisch universell eindeutig.** Er ist nur dann stabil, wenn Algorithmus, Unicode-Regeln und Duplicate-Handling fest definiert sind. Genau deshalb sind parser-/renderer-bewusste Tools wie **`md-toc`** oder **`mdformat-toc`** interessant.
- **Die meisten reifen Tools lösen nur ToC + Anchor-Verhalten**, aber **nicht** dein komplettes Zielbild aus **ToC + Überschriften-Nummerierung + Strip + State/Check**.
- Wenn du **möglichst nah an deinem geplanten `mdtoc`** bleiben willst, sind vor allem diese Kandidaten relevant:
  - **`kubernetes-sigs/mdtoc`**: sehr nah am kleinen Go-CLI-Gedanken
  - **`md-toc`**: am stärksten bei Renderer-/Slug-Kompatibilität
  - **`hoylen/markdown_toc`**: seltene Nummerierungs-Unterstützung, aber deutlich kleineres Projekt

## Nicht in der Haupttabelle, obwohl thematisch verwandt

- **BitDownToc**: bewusst **nicht** aufgenommen, weil der Maintainer im Repo die **Common Clause** als Lizenzentscheidung beschreibt; das ist für eine strenge „frei für OSS / offene Lizenz“-Auswahl kein sauberer Kandidat.
- **`gh-md-toc` (Shell)**: nicht in der Hauptliste, weil das Projekt selbst vor allem **Ubuntu/macOS** nennt und für **Windows** ausdrücklich auf die Go-Implementierung verweist.
- **`remark-toc`**: fachlich stark, aber eher **Plugin-Baustein** im `remark`-Ökosystem als direkter, kleiner `mdtoc`-Ersatz mit eigenem Marker-/Check-Workflow.
- **`codegourmet/markdown-toc`**: nicht aufgenommen, weil der Autor selbst von einem **„quick and dirty writeup“**, **fehlender Doku** und **fehlenden Tests** spricht.

## Kurzranking

1. **mdtoc (Kubernetes SIGs)** — 91/100
2. **doctoc** — 88/100
3. **md-toc** — 86/100
4. **mdformat-toc** — 82/100
5. **markdown-toc** — 78/100
6. **toc (ycd)** — 76/100
7. **markdown-toc-gen** — 74/100
8. **github-markdown-toc.go** — 69/100
9. **markdown_toc (Dart)** — 63/100

## Tabelle 1 — Tools waagerecht, Merkmale senkrecht

| Merkmal | mdtoc (Kubernetes SIGs) | doctoc | markdown-toc | md-toc | mdformat-toc | toc (ycd) | github-markdown-toc.go | markdown-toc-gen | markdown_toc (Dart) |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Empfehlung | 91 | 88 | 78 | 86 | 82 | 76 | 69 | 74 | 63 |
| Eigenschaften | Go-CLI; in-place, dry-run, stdin, glob patterns; GFM-only; Marker `<!-- toc --> ... <!-- /toc -->` | Node-CLI; auto-update zwischen Kommentaren; dry-run/stdout/update-only; git-hook-tauglich; Titel/Header/Footer; Min/Max-Level | Node-CLI und API; nutzt Remarkable; einfache ToC-Erzeugung; als Library gut integrierbar | Python-CLI/Library; offline; markerbasiert; Unterschieds-Check; mehrere Parser/Profile (GitHub, GitLab, CommonMark, Redcarpet …); pre-commit-Hook | Python-Plugin für `mdformat`; ToC an Markerzeile; GitHub-/GitLab-Slugfunktion; optionale HTML-Anker; Min/Max-Level | Go-CLI; `<!--toc-->` Marker; optional bulleted/numbered list; skip/depth; stdout oder append | Go-CLI; GitHub-spezifische ToC-Erzeugung; mehrere Dateien; Parallelverarbeitung; lokale/remote Dateien; Token-Unterstützung | TypeScript/Node-CLI; insert/update/dry-run/check; Batch-Globs; GFM- und pandoc-kompatible Navigation; ignoriert Codeblöcke | Dart-CLI/Package; generiert ToC und nummerierte Überschriften; Binär- oder Dart-Nutzung; Entfernen von ToC/Nummerierung möglich |
| Sprache | Go | JavaScript (Node.js) | JavaScript (Node.js) | Python | Python | Go | Go | TypeScript / Node.js | Dart |
| Lizenz | Apache-2.0 | MIT | MIT | GPL-3.0-or-later | MIT | Apache-2.0 | MIT | MIT | BSD-3-Clause |
| Interna | Single-purpose Go tool; normalized indentation; explicit tag-based update workflow | Comment-pragmas für Update; Marker-basierte In-place-Aktualisierung | Parser-/Library-orientiert; erzeugt Slugs/JSON auch per API | Profiler-/Parser-bewusste ToC-Regeln; Reverse-Engineering der Renderer-Regeln; viele Packaging-Kanäle | Plugin-Modell auf mdformat; standardmäßig HTML-Anker plus GitHub-Slug; sehr kleine Wheels | Zero-config-Ansatz; veröffentlichte Binaries für Windows/macOS/Linux | Go-Port von `gh-md-toc`; keine Shell-Dependencies; internetabhängig | Kleines CLI mit CI-orientiertem `check`; Fokus auf Prettier-konforme Ausgabe | Seltenes Tool mit Nummerierung + ToC; Pub-Paket und kompilierbares Executable |
| Download | [Repo](https://github.com/kubernetes-sigs/mdtoc) / [Releases](https://github.com/kubernetes-sigs/mdtoc/releases) | [Repo](https://github.com/thlorenz/doctoc) / [npm](https://www.npmjs.com/package/doctoc) | [Repo](https://github.com/jonschlinkert/markdown-toc) / [npm](https://www.npmjs.com/package/markdown-toc) | [Repo](https://github.com/frnmst/md-toc) / [PyPI](https://pypi.org/project/md-toc/) | [Repo](https://github.com/hukkin/mdformat-toc) / [PyPI](https://pypi.org/project/mdformat-toc/) | [Repo](https://github.com/ycd/toc) / [Releases](https://github.com/ycd/toc/releases) | [Repo](https://github.com/ekalinin/github-markdown-toc.go) / [Releases](https://github.com/ekalinin/github-markdown-toc.go/releases) | [Repo](https://github.com/thesilk-tux/markdown-toc-gen) / [npm](https://www.npmjs.com/package/markdown-toc-gen) | [Repo](https://github.com/hoylen/markdown_toc) / [Pub](https://pub.dev/packages/markdown_toc) |
| Review | [DeepWiki Overview](https://deepwiki.com/kubernetes-sigs/mdtoc/1-overview) | [DeepWiki](https://deepwiki.com/thlorenz/doctoc) / [LibHunt-Vergleich](https://www.libhunt.com/compare-doctoc-vs-vscode-markdown-pdf) | [LibHunt](https://www.libhunt.com/r/markdown-toc) / [Vergleich](https://www.libhunt.com/compare-markdown-toc-vs-github-markdown-toc) | [LibHunt](https://www.libhunt.com/r/frnmst/md-toc) / [Doku](https://docs.franco.net.eu.org/md-toc/) | [PyPI-Projektseite](https://pypi.org/project/mdformat-toc/) / [mdformat-Plugin-Doku](https://mdformat.readthedocs.io/en/stable/users/plugins.html) | [LibHunt-Vergleich](https://www.libhunt.com/compare-ycd--toc-vs-doctoc) | [Libraries.io](https://libraries.io/homebrew/githubmarkdowntoc) | [GitHub README](https://github.com/thesilk-tux/markdown-toc-gen) / [Release Notes](https://github.com/thesilk-tux/markdown-toc-gen/blob/master/RELEASE_NOTES.md) | [Pub-Paket](https://pub.dev/packages/markdown_toc) / [API-Doku](https://pub.dev/documentation/markdown_toc/latest/) |
| Verbreitung | 47 GitHub-Stars; 8 Releases; Kubernetes SIG Release-Kontext | 4.4k GitHub-Stars; npm-Paket mit 83 Dependents | 1.7k GitHub-Stars; npm meldet 205 weitere Projekte; offizielle README nennt u. a. NASA/openmct, Prisma, Prettier, Docusaurus | 38 GitHub-Stars; Suchsnippet nennt 405 dependent repos; Debian/nix/Anaconda/PyPI-Badges | 27 GitHub-Stars; Teil des mdformat-Ökosystems | 98 GitHub-Stars; 7 Releases | 522 GitHub-Stars; 22 Releases | 5 GitHub-Stars; 14 Tags; npm-Paket zuletzt 2025 aktualisiert | 0 GitHub-Stars; Pub-Metadaten zeigen 23 Downloads |
| OS | Go-basiert; Linux-Binary dokumentiert, `go install` verfügbar | multiplattform über Node.js; optional Docker-Workflow dokumentiert | multiplattform über Node.js | multiplattform über Python | multiplattform über Python >=3.9 | explizite Binaries für Windows 32/64, macOS 64, Linux 32/64/ARM64 | explizit cross-platform, inkl. Windows/Mac/Linux | multiplattform über Node.js | Dart-basiert; als Dart-Skript oder kompiliertes Executable nutzbar |
| Leichtgewicht | hoch – Go-CLI, standalone binary dokumentiert | mittel – npm global/local install; im CI sehr einfach | mittel – npm-Tool, aber sehr klein und gut skriptbar | mittel – Python-Runtime nötig, dafür offline und sehr flexibel | hoch – PyPI-Wheel ~9.7 kB, aber nur sinnvoll mit mdformat | sehr hoch – kleines Go-Tool mit direkter Binary-Nutzung | hoch – Go-Binary, keine externen Unix-Tools nötig | hoch – kleines npm-Tool, gute CI-Befehle | mittel – kleine Codebasis, aber Dart-Toolchain bzw. eigener Build nötig |
| GrosseNutzer | Kubernetes SIG Release / Kubernetes-Dokumentationsumfeld | breite OSS-Nutzung; viele Repo-/README-Workflows | NASA/openmct, Prisma, Prettier, Docusaurus, Joi, Mocha u. a. | laut Projekt vor allem Blogs/Docs/README-Workflows; 405 abhängige Repos im Badge/Snippet | keine prominent genannten Einzelprojekte; profitiert vom mdformat-Ökosystem | keine prominent genannten Einzelprojekte | starke GitHub-README/Wiki-Nutzung; aber GitHub-zentriert | keine prominent genannten Einzelprojekte | keine prominent genannten Einzelprojekte |
| Urteil | Bester direkter Ersatz, wenn du ein kleines, deterministisches CLI mit CI-Check willst; Nachteil: nur GFM und keine Heading-Nummerierung. | Sehr robuster ToC-Generator mit reifem CLI; sehr guter Ersatz, wenn Nummerierung der Überschriften nicht zwingend ist. | Starker Referenzkandidat für API/CLI und Verbreitung; schwächer als Ersatz für strenge Marker-/Check-Workflows. | Für renderer-kompatible Slugs/Links einer der besten Kandidaten; hervorragend, wenn Verhaltenskompatibilität wichtiger ist als Minimalismus. | Technisch sauber und klein; sehr gut, wenn ihr ohnehin mdformat nutzt oder Formatierung + ToC koppeln wollt. | Schlank und portabel; guter Ersatz für simple ToC-Workflows, aber deutlich schlichter als dein Zielbild. | Interessant, wenn Trice streng GitHub-kompatible Anchors braucht; als generischer `mdtoc`-Ersatz wegen Internetabhängigkeit nur zweite Wahl. | Sehr brauchbar für moderne CI-Pipelines; geringe Verbreitung ist der größte Nachteil. | Wegen Nummerierung hochrelevant für deine Spezifikation; wegen sehr kleiner Verbreitung und dokumentierter Codeblock-Limitation aber kein sorgenfreier Standardersatz. |

## Tabelle 2 — Tools senkrecht, Merkmale waagerecht

| Tool | Empfehlung | Eigenschaften | Sprache | Lizenz | Interna | Download | Review | Verbreitung | OS | Leichtgewicht | GrosseNutzer | Urteil |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| mdtoc (Kubernetes SIGs) | 91 | Go-CLI; in-place, dry-run, stdin, glob patterns; GFM-only; Marker `<!-- toc --> ... <!-- /toc -->` | Go | Apache-2.0 | Single-purpose Go tool; normalized indentation; explicit tag-based update workflow | [Repo](https://github.com/kubernetes-sigs/mdtoc) / [Releases](https://github.com/kubernetes-sigs/mdtoc/releases) | [DeepWiki Overview](https://deepwiki.com/kubernetes-sigs/mdtoc/1-overview) | 47 GitHub-Stars; 8 Releases; Kubernetes SIG Release-Kontext | Go-basiert; Linux-Binary dokumentiert, `go install` verfügbar | hoch – Go-CLI, standalone binary dokumentiert | Kubernetes SIG Release / Kubernetes-Dokumentationsumfeld | Bester direkter Ersatz, wenn du ein kleines, deterministisches CLI mit CI-Check willst; Nachteil: nur GFM und keine Heading-Nummerierung. |
| doctoc | 88 | Node-CLI; auto-update zwischen Kommentaren; dry-run/stdout/update-only; git-hook-tauglich; Titel/Header/Footer; Min/Max-Level | JavaScript (Node.js) | MIT | Comment-pragmas für Update; Marker-basierte In-place-Aktualisierung | [Repo](https://github.com/thlorenz/doctoc) / [npm](https://www.npmjs.com/package/doctoc) | [DeepWiki](https://deepwiki.com/thlorenz/doctoc) / [LibHunt-Vergleich](https://www.libhunt.com/compare-doctoc-vs-vscode-markdown-pdf) | 4.4k GitHub-Stars; npm-Paket mit 83 Dependents | multiplattform über Node.js; optional Docker-Workflow dokumentiert | mittel – npm global/local install; im CI sehr einfach | breite OSS-Nutzung; viele Repo-/README-Workflows | Sehr robuster ToC-Generator mit reifem CLI; sehr guter Ersatz, wenn Nummerierung der Überschriften nicht zwingend ist. |
| markdown-toc | 78 | Node-CLI und API; nutzt Remarkable; einfache ToC-Erzeugung; als Library gut integrierbar | JavaScript (Node.js) | MIT | Parser-/Library-orientiert; erzeugt Slugs/JSON auch per API | [Repo](https://github.com/jonschlinkert/markdown-toc) / [npm](https://www.npmjs.com/package/markdown-toc) | [LibHunt](https://www.libhunt.com/r/markdown-toc) / [Vergleich](https://www.libhunt.com/compare-markdown-toc-vs-github-markdown-toc) | 1.7k GitHub-Stars; npm meldet 205 weitere Projekte; offizielle README nennt u. a. NASA/openmct, Prisma, Prettier, Docusaurus | multiplattform über Node.js | mittel – npm-Tool, aber sehr klein und gut skriptbar | NASA/openmct, Prisma, Prettier, Docusaurus, Joi, Mocha u. a. | Starker Referenzkandidat für API/CLI und Verbreitung; schwächer als Ersatz für strenge Marker-/Check-Workflows. |
| md-toc | 86 | Python-CLI/Library; offline; markerbasiert; Unterschieds-Check; mehrere Parser/Profile (GitHub, GitLab, CommonMark, Redcarpet …); pre-commit-Hook | Python | GPL-3.0-or-later | Profiler-/Parser-bewusste ToC-Regeln; Reverse-Engineering der Renderer-Regeln; viele Packaging-Kanäle | [Repo](https://github.com/frnmst/md-toc) / [PyPI](https://pypi.org/project/md-toc/) | [LibHunt](https://www.libhunt.com/r/frnmst/md-toc) / [Doku](https://docs.franco.net.eu.org/md-toc/) | 38 GitHub-Stars; Suchsnippet nennt 405 dependent repos; Debian/nix/Anaconda/PyPI-Badges | multiplattform über Python | mittel – Python-Runtime nötig, dafür offline und sehr flexibel | laut Projekt vor allem Blogs/Docs/README-Workflows; 405 abhängige Repos im Badge/Snippet | Für renderer-kompatible Slugs/Links einer der besten Kandidaten; hervorragend, wenn Verhaltenskompatibilität wichtiger ist als Minimalismus. |
| mdformat-toc | 82 | Python-Plugin für `mdformat`; ToC an Markerzeile; GitHub-/GitLab-Slugfunktion; optionale HTML-Anker; Min/Max-Level | Python | MIT | Plugin-Modell auf mdformat; standardmäßig HTML-Anker plus GitHub-Slug; sehr kleine Wheels | [Repo](https://github.com/hukkin/mdformat-toc) / [PyPI](https://pypi.org/project/mdformat-toc/) | [PyPI-Projektseite](https://pypi.org/project/mdformat-toc/) / [mdformat-Plugin-Doku](https://mdformat.readthedocs.io/en/stable/users/plugins.html) | 27 GitHub-Stars; Teil des mdformat-Ökosystems | multiplattform über Python >=3.9 | hoch – PyPI-Wheel ~9.7 kB, aber nur sinnvoll mit mdformat | keine prominent genannten Einzelprojekte; profitiert vom mdformat-Ökosystem | Technisch sauber und klein; sehr gut, wenn ihr ohnehin mdformat nutzt oder Formatierung + ToC koppeln wollt. |
| toc (ycd) | 76 | Go-CLI; `<!--toc-->` Marker; optional bulleted/numbered list; skip/depth; stdout oder append | Go | Apache-2.0 | Zero-config-Ansatz; veröffentlichte Binaries für Windows/macOS/Linux | [Repo](https://github.com/ycd/toc) / [Releases](https://github.com/ycd/toc/releases) | [LibHunt-Vergleich](https://www.libhunt.com/compare-ycd--toc-vs-doctoc) | 98 GitHub-Stars; 7 Releases | explizite Binaries für Windows 32/64, macOS 64, Linux 32/64/ARM64 | sehr hoch – kleines Go-Tool mit direkter Binary-Nutzung | keine prominent genannten Einzelprojekte | Schlank und portabel; guter Ersatz für simple ToC-Workflows, aber deutlich schlichter als dein Zielbild. |
| github-markdown-toc.go | 69 | Go-CLI; GitHub-spezifische ToC-Erzeugung; mehrere Dateien; Parallelverarbeitung; lokale/remote Dateien; Token-Unterstützung | Go | MIT | Go-Port von `gh-md-toc`; keine Shell-Dependencies; internetabhängig | [Repo](https://github.com/ekalinin/github-markdown-toc.go) / [Releases](https://github.com/ekalinin/github-markdown-toc.go/releases) | [Libraries.io](https://libraries.io/homebrew/githubmarkdowntoc) | 522 GitHub-Stars; 22 Releases | explizit cross-platform, inkl. Windows/Mac/Linux | hoch – Go-Binary, keine externen Unix-Tools nötig | starke GitHub-README/Wiki-Nutzung; aber GitHub-zentriert | Interessant, wenn Trice streng GitHub-kompatible Anchors braucht; als generischer `mdtoc`-Ersatz wegen Internetabhängigkeit nur zweite Wahl. |
| markdown-toc-gen | 74 | TypeScript/Node-CLI; insert/update/dry-run/check; Batch-Globs; GFM- und pandoc-kompatible Navigation; ignoriert Codeblöcke | TypeScript / Node.js | MIT | Kleines CLI mit CI-orientiertem `check`; Fokus auf Prettier-konforme Ausgabe | [Repo](https://github.com/thesilk-tux/markdown-toc-gen) / [npm](https://www.npmjs.com/package/markdown-toc-gen) | [GitHub README](https://github.com/thesilk-tux/markdown-toc-gen) / [Release Notes](https://github.com/thesilk-tux/markdown-toc-gen/blob/master/RELEASE_NOTES.md) | 5 GitHub-Stars; 14 Tags; npm-Paket zuletzt 2025 aktualisiert | multiplattform über Node.js | hoch – kleines npm-Tool, gute CI-Befehle | keine prominent genannten Einzelprojekte | Sehr brauchbar für moderne CI-Pipelines; geringe Verbreitung ist der größte Nachteil. |
| markdown_toc (Dart) | 63 | Dart-CLI/Package; generiert ToC und nummerierte Überschriften; Binär- oder Dart-Nutzung; Entfernen von ToC/Nummerierung möglich | Dart | BSD-3-Clause | Seltenes Tool mit Nummerierung + ToC; Pub-Paket und kompilierbares Executable | [Repo](https://github.com/hoylen/markdown_toc) / [Pub](https://pub.dev/packages/markdown_toc) | [Pub-Paket](https://pub.dev/packages/markdown_toc) / [API-Doku](https://pub.dev/documentation/markdown_toc/latest/) | 0 GitHub-Stars; Pub-Metadaten zeigen 23 Downloads | Dart-basiert; als Dart-Skript oder kompiliertes Executable nutzbar | mittel – kleine Codebasis, aber Dart-Toolchain bzw. eigener Build nötig | keine prominent genannten Einzelprojekte | Wegen Nummerierung hochrelevant für deine Spezifikation; wegen sehr kleiner Verbreitung und dokumentierter Codeblock-Limitation aber kein sorgenfreier Standardersatz. |

## Leseschlüssel für das Empfehlungslevel

- **90–100**: sehr guter Ersatzkandidat
- **80–89**: stark, aber mit klarer Einschränkung
- **70–79**: brauchbar, wenn das Toolprofil zu deinem Workflow passt
- **60–69**: nur für spezielle Randfälle interessant
- **<60**: eher Inspirationsquelle als echter Ersatz

## Mein Fazit

Wenn du **ein kleines, sauberes, plattformtaugliches CLI-Vorbild** suchst, ist **`kubernetes-sigs/mdtoc`** die beste Referenz.  
Wenn du **Renderer-/Slug-Kompatibilität** priorisierst, ist **`frnmst/md-toc`** inhaltlich am interessantesten.  
Wenn dir **Nummerierung von Überschriften** wichtig ist, ist **`hoylen/markdown_toc`** relevant – aber eher als Ideenlieferant als als Standardwerkzeug.