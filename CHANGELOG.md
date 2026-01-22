# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
-->

## [0.2.1][] - 2026-01-21

### Changed

* `make` command now uses original text (msgid) as fallback when translation
  is empty, has `notranslate` flag, or entry is missing in PO file.

[0.2.1]: https://github.com/WoozyMasta/dayz-stringtable/compare/v0.2.0...v0.2.1

## [0.2.0][] - 2026-01-21

### Added

* `stats` command for translation statistics with completion percentages,
  verbose mode, and JSON output format
* `clean` command to remove duplicate `msgstr` entries and unused keys
  from PO files
* Full PO file header support with automatic updates and content-based
  change detection
* Complete comment support (translator, extracted, reference, flag) with
  preservation on updates
* `# notranslate` flag handling for intentionally untranslated entries
* `LoadPODirectory()` utility function for batch loading PO files
* Makefile for build automation
* Refactored packages: `internal/csvutil` and `internal/poutil`

### Changed

* **Fixed PO file sorting**: Entries now follow CSV order
  (was an error causing noisy commits)
* `update` command now preserves all comments from existing entries
* Improved error handling: functions now return errors instead of exiting
* Project structure refactored to modular packages
* Build system migrated to Makefile

### Note

Commits after v0.1.2 may appear noisy due to the sorting fix.
This correction ensures CSV and PO file ordering consistency.

[0.2.0]: https://github.com/WoozyMasta/dayz-stringtable/compare/v0.1.2...v0.2.0

## [0.1.2][] - 2025-07-16

### Added

* Automatic check for duplicate keys in the first column of CSV (msgctxt).
  Now `LoadCSV()` will return an error if any duplicates are found.

[0.1.2]: https://github.com/WoozyMasta/dayz-stringtable/compare/v0.1.1...v0.1.2

## [0.1.1][] - 2025-05-25

### Added

* Usage examples
* `.markdownlint.json` configuration

## Changed

* Fixed behavior of update command, new and untranslated translation
  strings are now empty in the resulting po files as expected
* Disabled UPX compression for windows builds due to large number of
  false positives on VirusTotal

[0.1.1]: https://github.com/WoozyMasta/dayz-stringtable/compare/v0.1.0...v0.1.1

## [0.1.0][] - 2025-05-24

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/dayz-stringtable/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
