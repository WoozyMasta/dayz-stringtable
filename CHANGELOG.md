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
