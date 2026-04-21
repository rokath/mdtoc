# Install Smoke Fixture

This fixture is used by the install-check workflow.

Expected smoke-test signals after `mdtoc generate`:

* `+` becomes the managed ToC bullet style because live document bullets use `+` most often
* repeated headings produce distinct anchors such as `#overview` and `#overview-1`
* headings with leading numbers keep those digits in generated anchors
* headings inside fenced code blocks stay ignored
* headings inside `<!-- mdtoc off -->` ... `<!-- mdtoc on -->` stay unmanaged

+ dominant plus bullet
+ another dominant plus bullet
+ third dominant plus bullet
+ fourth dominant plus bullet

## 2026 Release Plan

Intro paragraph for the numbered heading case.

### Scope

+ nested plus bullet
+ another nested plus bullet

## Overview

```md
## Code Heading
+ code bullet
```

<!-- mdtoc off -->
## Hidden Section
- excluded dash bullet
### Hidden Details
<!-- mdtoc on -->

## Overview

## Final Notes
