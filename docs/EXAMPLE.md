# Example Markdown File

## 1. Examples for ignored Headings

### 1.1. Examples as Setext headings

mdtoc ignores this as heading for ToC generation:

IGNORED Heading 1
=================

IGNORED Heading 2
-----------------

### 1.2. Examples in fenced code blocks

mdtoc ignores this as heading for ToC generation:

```md
## IGNORED Heading 3
```

````md
``` 
## IGNORED Heading 4
```
````

### 1.3. Example in HTML comments

mdtoc ignores this as heading for ToC generation: (invisible)

<!-- 
## IGNORED Heading 5
-->

### 1.4. Example as HTML syntax

mdtoc ignores this as heading for ToC generation:

<h6>IGNORED Heading 6</h6>

### 1.5. Example inside exclusion region

mdtoc ignores this as heading for ToC generation:

<!-- mdtoc off -->
###### IGNORED Heading 7
<!-- mdtoc on -->

### 1.6. Examples with starting space(s)

mdtoc takes this as heading for ToC generation:

#### 1.6.1. 0 starting space Example NOT ignored

mdtoc ignores this as heading for ToC generation:

 #### IGNORED 1 starting space Example

## 2. Example for closed ATX heading    ########

If you generate with `anchor=false` (headlines without link anchors), the right `slug` value depends on the used Markdown renderer. For example with _VS Code Markdown Enhanced_ extension, then `slug=crossnote` is important to get working ToC links for closed ATX headings. With `anchor=true` (default), the `slug` value does not really matter, because the link anchors in the headings match automatically the ToC links.

## 3. Examples for repeated Headings

Users should not publish generated link anchors to repeated headings. See closed issue [#8](https://github.com/rokath/mdtoc/issues/8) for details. Instead they could publish manually created links, like `<a id="chapter-a-about"></a>`. But it is users responsibility not to create the same link anchor twice. See closed issue [#97](https://github.com/rokath/mdtoc/issues/97) for details.

### 3.1. Chapter A

#### 3.1.1. About

<a id="chapter-a-about"></a>

### 3.2. Chapter B

#### 3.2.1. About

### 3.3. Chapter C

#### 3.3.1. About

<img src="../extension/mdtoc_mascot_1024.webp" width="400">

<details markdown="1"> <!-- parse this block as markdown -->
<summary><strong style="font-size: 1.25em;">Table of Contents</strong> <span style="font-size: 0.66em;">(click to expand)</span></summary>

<!-- mdtoc -->

* [1. Examples for ignored Headings](#1-examples-for-ignored-headings)
  * [1.1. Examples as Setext headings](#11-examples-as-setext-headings)
  * [1.2. Examples in fenced code blocks](#12-examples-in-fenced-code-blocks)
  * [1.3. Example in HTML comments](#13-example-in-html-comments)
  * [1.4. Example as HTML syntax](#14-example-as-html-syntax)
  * [1.5. Example inside exclusion region](#15-example-inside-exclusion-region)
  * [1.6. Examples with starting space(s)](#16-examples-with-starting-spaces)
    * [1.6.1. 0 starting space Example NOT ignored](#161-0-starting-space-example-not-ignored)
* [2. Example for closed ATX heading](#2-example-for-closed-atx-heading----)
* [3. Examples for repeated Headings](#3-examples-for-repeated-headings)
  * [3.1. Chapter A](#31-chapter-a)
    * [3.1.1. About](#311-about)
  * [3.2. Chapter B](#32-chapter-b)
    * [3.2.1. About](#321-about)
  * [3.3. Chapter C](#33-chapter-c)
    * [3.3.1. About](#331-about)

<!-- numbering=true min=2 max=4 slug=github anchor=false link=true toc=true bullets=auto -->
<!-- /mdtoc -->

</details>

---
