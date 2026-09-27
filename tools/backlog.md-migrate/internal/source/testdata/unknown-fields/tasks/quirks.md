---
id: FIX-20
title: Task with a quirk the CLI cannot produce on its own
status: To Do
assignee: []
created_date: '2026-09-26 22:50'
labels: []
dependencies: []
due_date: '01/10/2026'
custom_field: something the tool has never seen
ordinal: 20000
---

<!--
  Hand-edited fixture: the CLI never writes an unrecognized frontmatter key,
  an unrecognized body section, an invalid due_date shape, or a comment
  block missing "created:" (backlog task edit --comment always stamps one).
  Every other testdata fixture in this repository is generated verbatim by
  the real Backlog.md CLI; this file exists only to exercise the findings
  that a hand-crafted, malformed-but-parseable input can trigger, which the
  CLI itself refuses to write.
-->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Something to check
<!-- AC:END -->

## Review

<!-- SECTION:REVIEW:BEGIN -->
A section the specification does not know about.
<!-- SECTION:REVIEW:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: @ann
---
A comment with no created line at all.
---
<!-- COMMENTS:END -->
