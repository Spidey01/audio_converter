# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- markdownlint-disable MD024 -->

## [Unreleased]

This release now allows control over the conversion profile when running the
exporter. Additionally, the individual exporters now have flags to convert and
scale cover art. The behavior of options across programs and how they are tested
has been further normalized.

Also, there is now a change log :).

### Changed

- All programs now support the `-version`, `-log-file`, `-n`, `-y`, and `-v` flags.
- export_audio_tree
  - Output can now be controlled using the same flags as to_aac, to_flac, etc.
- extract_coverart added `-cover` and `-scale` as aliases for `-c` and `-s`

### Added

- to_aac, to_flac, to_mp3
  - Added `-cover` flag to specify how to convert cover art. Default is "copy" to maintain original behavior.
  - Added `-scale` flag to specify size of cover art, when `-cover` specifies a conversion.

### Fixed

- Getting the version no longer prints an error on startup when run from `$PATH`.

## [v1.1.0] - 2025-08-19

This release converts export_audio_tree from a single-threaded synchronous
operation to running multiple Goroutines in a work pool for copying / converting
files. On my system with 8 cores, this takes a Flac export of my library from
about 30-40 minutes down to about 5-10 minutes with the default settings.

### Changed

- export_audio_tree now uses a workpool.

### Added

- extract_coverart now takes a [`-c codec`] option.
- export_audio_tree
  - Added `-y` flag to overwrite files instead of prompting.
  - Added `-q` and `-j` flags to control the work pool.

### Fixed

- extract_coverart copying rather than converting.
- Fixed several unit tests.

## [v1.0.0] - 2025-08-16

Initial release. This represents the crossing point between what my original
bash scripts could almost do and the rewrite in Go becoming feature complete.
So, this is from prototype to it works!

This first release supports conversion of individual files to AAC, FLAC, and
MP3, along with tools for exporting directory trees and extracting cover art.

---

- \[unreleased\]: [changes](https://github.com/Spidey01/audio_converter/compare/v1.1.0...HEAD)
- [[1.1.0](https://github.com/Spidey01/audio_converter/releases/tag/v1.1.0)]: [changes](https://github.com/Spidey01/audio_converter/compare/v1.0.0...v1.1.0)
- [[1.0.0](https://github.com/Spidey01/audio_converter/releases/tag/v1.0.0)]: [commits](https://github.com/Spidey01/audio_converter/commits/v1.0.0/)
