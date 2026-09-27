---
id: FIX-40
title: Description section marker was never closed
status: To Do
assignee: []
labels: []
dependencies: []
ordinal: 40000
---

<!--
  Hand-edited fixture: the Backlog.md CLI always writes a "BEGIN" marker
  together with its matching "END", so it never leaves one unclosed on its
  own. This file removes the closing "<!-- SECTION:DESCRIPTION:END -->" of
  an otherwise normal task body to exercise the "was never closed" finding.
  The sections around it (Acceptance Criteria, Implementation Plan) stay
  well formed and must still be read normally, unaffected by the broken
  Description.
-->

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
This text is never reached: the closing marker was removed on purpose, so
the whole section is treated as absent.

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 First criterion
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
A short plan that must still be read even though Description above it is
broken.
<!-- SECTION:PLAN:END -->
