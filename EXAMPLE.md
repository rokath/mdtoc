# Example Marfkdown File

<img src="./extension/mdtoc_mascot_1024.webp" width="200">

## 1. <a id="about"></a>About

## 2. <a id="examples-for-ignored-headings"></a>Examples for ignored Headings

### 2.1. <a id="examples-as-setext-headings"></a>Examples as Setext headings

IGNORED Example 1
=================

IGNORED Example 2
-----------------

### 2.2. <a id="examples-in-fenced-code-blocks"></a>Examples in fenced code blocks

```md
## IGNORED Example in fenced code block
```

````md
``` 
## IGNORED Example in fenced code block inside fenced code block
```
````

### 2.3. <a id="example-in-html-comments"></a>Example in HTML comments

<!-- 
## IGNORED HTML comment Example
-->

### 2.4. <a id="example-as-html-syntax"></a>Example as HTML syntax

<h2>IGNORED HTML Example</h2>

### 2.5. <a id="example-inside-exclusion-region"></a>Example inside exclusion region

<!-- mdtoc off -->
## IGNORED Example mdtoc off
<!-- mdtoc on -->

### 2.6. <a id="examples-with-starting-space-s"></a>Examples with starting space(s)

#### 2.6.1. <a id="0-starting-space-example-not-ignored"></a>0 starting space Example NOT ignored
 
 #### IGNORED 1 starting space Example

## 3. <a id="footnotes"></a>Footnotes ##

<h2>Table of Contents</h2>

<!-- mdtoc -->
* [1. About](#about)
* [2. Examples for ignored Headings](#examples-for-ignored-headings)
  * [2.1. Examples as Setext headings](#examples-as-setext-headings)
  * [2.2. Examples in fenced code blocks](#examples-in-fenced-code-blocks)
  * [2.3. Example in HTML comments](#example-in-html-comments)
  * [2.4. Example as HTML syntax](#example-as-html-syntax)
  * [2.5. Example inside exclusion region](#example-inside-exclusion-region)
  * [2.6. Examples with starting space(s)](#examples-with-starting-space-s)
    * [2.6.1. 0 starting space Example NOT ignored](#0-starting-space-example-not-ignored)
* [3. Footnotes ##](#footnotes)
<!-- mdtoc-config
container-version=v2
numbering=true
min-level=2
max-level=4
anchor=github
toc=true
bullets=auto
state=generated
-->
<!-- /mdtoc -->

