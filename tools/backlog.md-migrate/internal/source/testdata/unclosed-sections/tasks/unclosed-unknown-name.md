---
id: FIX-41
title: Unknown section marker was never closed
status: To Do
assignee: []
labels: []
dependencies: []
ordinal: 41000
---

<!--
  Hand-edited fixture: same reasoning as unclosed-description.md, but for a
  marker name the specification does not know at all ("REVIEW", without the
  "SECTION:" prefix). It must raise both the existing "unrecognized body
  section" finding (the name is not in the mapping table) and the new "was
  never closed" finding (its BEGIN has no END anywhere later in the body):
  the two findings describe different problems and neither one replaces
  the other. The Description section around it is well formed and must
  still be read normally.
-->

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A normal description, unaffected by the broken Review section below.
<!-- SECTION:DESCRIPTION:END -->

## Review

<!-- REVIEW:BEGIN -->
This text is never reached: there is no closing marker for this section
anywhere below.
