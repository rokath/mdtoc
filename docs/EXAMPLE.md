# Example Markdown File

## Examples for ignored Headings

### Examples as Setext headings

IGNORED Heading 1
=================

IGNORED Heading 2
-----------------

### Examples in fenced code blocks

```md
## IGNORED Heading 3
```

````md
``` 
## IGNORED Heading 4
```
````

### Example in HTML comments

<!-- 
## IGNORED Heading 5
-->

### Example as HTML syntax

<h6>IGNORED Heading 6</h6>

### Example inside exclusion region

<!-- mdtoc off -->
###### IGNORED Heading 7
<!-- mdtoc on -->

### Examples with starting space(s)

#### 0 starting space Example NOT ignored
 
 #### IGNORED 1 starting space Example

## Example for closed ATX heading    ########

If you generate with `anchor=false` (headlines without link anchors), the right `slug` value depends on the used Markdown renderer. For example with _VS Code Markdown Enhanced_ extension, then `slug=crossnote` is important to get working ToC links for closed ATX headings. With `anchor=true` (default), the `slug` value does not really matter, because the link anchors in the headings match automatically the ToC links.

## Examples for repeated Headings

Users should not publish generated link anchors to repeated headings. See closed issue [#8](https://github.com/rokath/mdtoc/issues/8) for details. Instead they could publish manually created links, like `<a id="chapter-a-about"></a>`.

### Chapter A

#### About

<a id="chapter-a-about"></a>

### Chapter B

#### About

### Chapter C

#### About

<img src="../extension/mdtoc_mascot_1024.webp" width="400">

<h2>Table of Contents</h2>

<!-- mdtoc -->
<!-- numbering=true min=2 max=4 slug=github anchor=false link=true toc=true bullets=auto -->
<!-- /mdtoc -->
