# mdtoc VS Code Extension: Executable Policy

## Ziel

Die VS-Code-Extension soll für normale User ohne zusätzliche Installation funktionieren und gleichzeitig Power-Usern die volle Kontrolle über ein eigenes `mdtoc`-Binary geben.

## Grundsatz

Die Extension kennt genau zwei Betriebsarten:

1. **Bundled Mode**
2. **Custom Executable Mode**

Es gibt keinen automatischen `PATH`-Fallback und keine automatische Auswahl eines systemweit installierten `mdtoc`.

## Bundled Mode

Wenn kein eigener Pfad konfiguriert ist, muss die Extension immer das mitgelieferte `mdtoc`-Binary verwenden.

Das gilt auch dann, wenn auf dem System zusätzlich ein `mdtoc` im `PATH` vorhanden ist.

```text
customPath leer
→ bundled mdtoc verwenden
```

Der Bundled Mode ist der Default.

## Custom Executable Mode

Wenn der User einen Pfad zu einem `mdtoc`-Binary konfiguriert, muss die Extension exakt dieses Binary verwenden.

```text
customPath gesetzt
→ exakt dieses Binary verwenden
```

Der konfigurierte Pfad bleibt bei Extension-Updates erhalten.

Der User ist in diesem Modus selbst dafür verantwortlich, das externe `mdtoc`-Binary aktuell und kompatibel zu halten.

## Kein automatischer PATH-Zugriff

Die Extension darf nicht automatisch nach `mdtoc` im System-`PATH` suchen.

Insbesondere ist folgende Logik nicht erlaubt:

```text
PATH → bundled
customPath → PATH → bundled
```

Erlaubt ist ausschließlich:

```text
customPath → bundled
```

## CLI-Verhalten

Das gebündelte `mdtoc`-Binary wird nur intern von der VS-Code-Extension verwendet.

Die Extension darf das Bundle-Binary nicht automatisch in den System-`PATH` eintragen.

Auf der Kommandozeile wird weiterhin nur das systemweit installierte oder explizit adressierte `mdtoc` verwendet.

## Setting

Die Extension soll genau ein User-Setting für den Override anbieten:

```json
{
  "mdtoc.executable.customPath": ""
}
```

Bedeutung:

```text
leer   = bundled mdtoc verwenden
gesetzt = angegebenes Binary verwenden
```

Der Pfad sollte als maschinenbezogenes Setting behandelt werden, nicht als normales projektweites Setting.

## Warnung bei Custom Path

Wenn ein Custom Path gesetzt wird oder erstmals verwendet wird, soll die Extension den User warnen:

```text
You configured a custom mdtoc executable.
The extension will use this binary instead of the bundled one.
You are responsible for keeping it compatible and up to date.
```

Der User soll die Möglichkeit haben, den Custom Path beizubehalten oder zum Bundled Mode zurückzukehren.

## Versionsprüfung

Die Extension soll `mdtoc --version` ausführen und die gefundene Version prüfen.

Für Custom Executable Mode gilt:

- Ist die Version kompatibel, darf sie verwendet werden.
- Ist die Version zu alt oder nicht ermittelbar, soll die Extension eine klare Fehlermeldung anzeigen.
- Die Extension darf in diesem Fall nicht still auf das bundled Binary zurückfallen.

## Updates

Im Bundled Mode wird `mdtoc` durch Updates der VS-Code-Extension aktualisiert.

Im Custom Executable Mode werden Extension und externes `mdtoc` getrennt aktualisiert:

```text
Extension Update → aktualisiert die Extension
Custom mdtoc Update → Verantwortung des Users
```

Die Extension soll kein eigenes Auto-Update für externe oder gebündelte `mdtoc`-Binaries implementieren.

## Diagnose

Die Extension soll einen Befehl bereitstellen:

```text
mdtoc: Show Version
```

Dieser Befehl soll mindestens anzeigen:

```text
VS Code extension version
mdtoc mode: bundled | custom
mdtoc path
mdtoc version
bundled mdtoc version, falls custom aktiv ist
```

## Normative Resolver-Logik

Die zentrale Resolver-Funktion muss sinngemäß so arbeiten:

```text
if mdtoc.executable.customPath is not empty:
    use customPath
else:
    use bundled binary for current platform
```

Alle Extension-Kommandos müssen diese zentrale Resolver-Logik verwenden.

## Zusammenfassung

```text
Default für normale User:
Bundled mdtoc

Override für Power-User:
Expliziter Custom Path

Nicht erlaubt:
Automatische PATH-Erkennung
Stiller Fallback
Separates Binary-Auto-Update
Automatisches Eintragen des Bundles in PATH
```
