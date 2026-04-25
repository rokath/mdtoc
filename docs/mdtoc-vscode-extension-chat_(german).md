# mdtoc als VS-Code-Extension

## Ausgangsfrage

Thomas Höhenleitner fragte sinngemäß:

> Ich habe OSS <https://github.com/rokath/mdtoc> kreiert. Wenn dieses derzeit in Go geschriebene Tool zukünftig auch als VS-Code-Extension verfügbar sein soll, ggf. mit sinnvollen Einschränkungen, was ist dazu nötig? Muss es anders codiert werden? Derzeit ist es in Go umgesetzt. Das Tool als CLI soll unbedingt erhalten bleiben.

## Kernaussage

`mdtoc` muss nicht grundsätzlich nach TypeScript oder JavaScript portiert werden. Die bestehende Go-CLI kann erhalten bleiben und als eigentliche Engine dienen. Die VS-Code-Extension wäre dann ein dünner Adapter in TypeScript, der das Go-Binary aufruft, Einstellungen aus VS Code in CLI-Argumente übersetzt und Änderungen sauber im Editor anwendet.

Die empfohlene Architektur lautet daher:

```text
Go bleibt kanonisch.
CLI bleibt voll erhalten.
VS-Code-Extension ruft CLI auf.
Web-Support wird explizit als späteres oder limitiertes Ziel behandelt.
```

## Empfohlene Architektur

Eine sinnvolle Repository-Struktur könnte so aussehen:

```text
mdtoc repository
├─ cmd/mdtoc/              # bleibt die CLI
├─ internal/mdtoc/         # bestehende Engine/Runner
├─ extension/              # neue VS-Code-Extension in TypeScript
│  ├─ package.json
│  ├─ src/extension.ts
│  └─ bin/<platform>/mdtoc
└─ .github/workflows/...
```

Die VS-Code-Extension sollte nicht die Markdown-Logik neu implementieren, sondern `mdtoc` per `child_process.execFile` oder ähnlichem aufrufen.

## Warum die Go-CLI erhalten bleiben kann

Der bestehende Ansatz ist für eine Extension günstig, weil `mdtoc` bereits als CLI funktioniert. Eine Extension kann das aktive Markdown-Dokument lesen, den Inhalt an `mdtoc` über `stdin` übergeben und das Ergebnis aus `stdout` zurück in den Editor schreiben.

Vorteile:

- Die Go-Implementierung bleibt die einzige kanonische Implementierung.
- CLI, CI und Editor liefern dasselbe Verhalten.
- Es entsteht keine zweite Markdown-/TOC-Implementierung in TypeScript.
- Das bestehende CLI-Tool bleibt weiterhin separat nutzbar.
- Fehler können über `stderr` und Exit Codes sauber an VS Code gemeldet werden.

## Was für eine VS-Code-Extension nötig wäre

### 1. TypeScript-Extension als Adapter

Die Extension registriert Befehle wie:

```json
{
  "contributes": {
    "commands": [
      { "command": "mdtoc.generate", "title": "mdtoc: Generate ToC" },
      { "command": "mdtoc.regen", "title": "mdtoc: Regenerate ToC" },
      { "command": "mdtoc.strip", "title": "mdtoc: Strip ToC" },
      { "command": "mdtoc.check", "title": "mdtoc: Check ToC" }
    ],
    "menus": {
      "editor/title": [
        {
          "command": "mdtoc.regen",
          "when": "resourceLangId == markdown",
          "group": "navigation"
        }
      ],
      "editor/context": [
        {
          "command": "mdtoc.generate",
          "when": "resourceLangId == markdown"
        }
      ]
    }
  }
}
```

Damit erscheinen die Funktionen in der Command Palette und optional im Editor-Kontextmenü oder in der Editor-Toolbar.

### 2. Go-Binary aus der Extension heraus aufrufen

Für den MVP sollte die Extension den Inhalt des aktiven Dokuments an `mdtoc` übergeben:

```ts
execFile(mdtocPath, ["generate"], { input: documentText }, ...)
```

Praktisch ist ein stdin/stdout-Modell:

- aktuelles Dokument lesen,
- Inhalt an `mdtoc` senden,
- transformierten Markdown aus `stdout` lesen,
- Änderung per VS-Code-Edit anwenden.

Das ist besser als direkt die Datei vom Go-Tool überschreiben zu lassen, weil VS Code dann Undo, Dirty-State und Editor-Änderungen korrekt behandeln kann.

### 3. VS-Code-Settings auf CLI-Optionen abbilden

Sinnvolle Settings wären zum Beispiel:

```json
{
  "mdtoc.numbering": true,
  "mdtoc.anchors": true,
  "mdtoc.toc": true,
  "mdtoc.minLevel": 2,
  "mdtoc.maxLevel": 4,
  "mdtoc.bullets": "auto",
  "mdtoc.formatOnSave": false,
  "mdtoc.binaryPath": ""
}
```

Diese Settings würden dann in CLI-Flags oder Modi übersetzt.

### 4. Binary-Strategie

Es gibt zwei Hauptvarianten.

#### Variante A: vorhandenes `mdtoc` aus dem PATH nutzen

Vorteile:

- Extension bleibt klein.
- Keine Plattform-Binaries im Extension-Paket.
- CLI-Installation bleibt explizit.

Nachteile:

- Nutzer müssen `mdtoc` separat installieren.
- Fehlerfälle durch fehlenden PATH oder falsche Version sind wahrscheinlicher.

#### Variante B: `mdtoc`-Binaries mitliefern

Vorteile:

- Beste User Experience.
- Installation der Extension reicht aus.
- Version von Extension und CLI-Binary ist kontrolliert.

Nachteile:

- Builds für mehrere Plattformen nötig.
- Größere VSIX-Pakete.
- Release-Prozess wird etwas komplexer.

Empfohlene Hybrid-Lösung:

1. Wenn `mdtoc.binaryPath` gesetzt ist, diesen Pfad verwenden.
2. Sonst gebündeltes Binary verwenden.
3. Sonst `mdtoc` aus `PATH` probieren.
4. Sonst eine klare Fehlermeldung mit Installationshinweis anzeigen.

## Muss am Go-Code etwas geändert werden?

Für einen MVP: kaum.

Sinnvolle Verbesserungen wären aber:

### 1. Stabile machine-readable Schnittstelle

Hilfreich wären Optionen wie:

- `--output text|json`,
- eindeutige Exit Codes,
- transformierter Markdown ausschließlich auf `stdout`,
- Fehlermeldungen ausschließlich auf `stderr`.

### 2. Extension-freundliches `check`

Für VS Code wäre ein klarer Exit-Code-Vertrag hilfreich:

```text
0   Dokument ist OK
1   Dokument würde geändert werden / ToC ist veraltet
2+  echter Fehler
```

Optional könnte `check` strukturierte JSON-Ausgaben liefern, damit VS Code Diagnostics anzeigen kann.

### 3. Go-Modulpfad bereinigen

Falls das Projekt öffentlich als Go-Modul nutzbar sein soll, wäre statt eines Beispielpfads ein echter Modulpfad sinnvoll:

```text
module github.com/rokath/mdtoc
```

Das ist besonders relevant, falls später Teile als Go-Library, WASM-Modul oder externe API genutzt werden sollen.

### 4. `internal` nur bei Library-Ziel überdenken

Solange die Extension nur das Binary aufruft, kann die interne Struktur bleiben.

Falls die Kernlogik aber später von anderen Go-Projekten importiert werden soll, wäre ein öffentliches Package wie `pkg/mdtoc` sinnvoller.

## Sinnvolle Einschränkungen für eine erste Version

Für eine erste VS-Code-Version wären diese Einschränkungen sinnvoll:

- nur Desktop VS Code,
- kein `vscode.dev` oder `github.dev`,
- nur aktive Markdown-Datei,
- kein Workspace-weites Batch-Processing,
- kein Format-on-save per Default,
- keine Markdown-Preview-Erweiterung,
- keine Virtual-Workspace-Unterstützung im MVP,
- bestehende funktionale Limits von `mdtoc` beibehalten.

Der wichtigste Punkt ist Desktop-only: Eine Web-Extension kann kein natives Go-Binary starten. Für Web-Support müsste die Logik nach TypeScript portiert oder nach WASM gebracht werden.

## Wann wäre eine Portierung nötig?

Eine Portierung wäre nur nötig, wenn eines dieser Ziele wichtig wird:

1. **Support für `vscode.dev` oder `github.dev`**  
   Dann kann kein natives Go-Binary gestartet werden. Die Logik müsste in TypeScript laufen oder als WASM verfügbar sein.

2. **Sehr tiefe VS-Code-Integration ohne Prozessaufruf**  
   Zum Beispiel Live-Diagnostics bei jedem Tastendruck, Preview-Sync oder Code Actions mit präzisen Positionsdaten.

3. **Extension ohne Plattform-Binaries**  
   Dann wäre ein TypeScript-Core oder ein WASM-Build attraktiver.

Für das Ziel „CLI unbedingt erhalten“ ist eine Portierung aber nicht erforderlich.

## Vorgeschlagene Roadmap

### Phase 1: MVP

- `extension/` anlegen.
- Commands implementieren:
  - `generate`,
  - `regen`,
  - `strip`,
  - `check`.
- Aktives Markdown-Dokument lesen.
- `mdtoc` per stdin/stdout aufrufen.
- Ergebnis per `WorkspaceEdit` oder TextEditor-Edit ersetzen.
- Fehler in einem VS-Code-Output-Channel anzeigen.
- Binary-Erkennung implementieren:
  - Setting,
  - gebündeltes Binary,
  - PATH.

### Phase 2: Distribution

- GitHub Actions oder GoReleaser um VS-Code-Builds erweitern.
- Platform-spezifische VSIX-Pakete bauen.
- Marketplace-Publisher einrichten.
- Extension README schreiben.
- Screenshots, Settings und Beispiele ergänzen.

### Phase 3: Komfort

- Editor Title Button für Markdown.
- Context Menu.
- Optional Format-on-save.
- `mdtoc check` als Diagnostic.
- Diff-Preview vor Änderung.
- Status-Bar-Anzeige wie `mdtoc OK` oder `ToC stale`.

### Phase 4: Optionaler Web-Support

- Prüfen, ob Go-Core nach WASM realistisch ist.
- Alternativ Kernlogik nach TypeScript portieren.
- Browser-spezifischen Extension-Entry ergänzen.

## Zusammenfassung

Die beste Lösung ist eine Mischarchitektur:

```text
VS-Code-Extension = TypeScript-Adapter
mdtoc = Go-Engine und CLI
```

Damit bleibt die CLI vollständig erhalten, während die Extension lediglich Editor-Integration, Settings, Menüs, Diagnostics und Distribution übernimmt.

Eine Neucodierung in TypeScript ist nicht nötig, solange die Extension nur für Desktop VS Code gedacht ist. Erst für Web-Extensions, sehr tiefe Live-Integration oder binary-freie Distribution wäre eine Portierung oder WASM-Variante relevant.
