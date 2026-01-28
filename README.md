# DayZ StringTable GetText CLI

A simple CLI tool for managing localization of DayZ `stringtables.csv`
and Gettext formats.

This utility helps you:

* **Generate a POT template** from your base CSV of original strings.
* **Create PO files** (per language) from CSV, with empty translations.
* **Merge PO files** back into a full CSV of translations.
* **Update existing PO** files when you add new strings to CSV.
* **Show translation statistics** with completion percentages
  and untranslated strings.

Under the hood it uses simplified [gettext] `.pot` and `.po` formats.

## About POT & PO

* `.pot`: template with all `msgid` (original texts) and `msgctxt` (keys),
  empty `msgstr`.
* `.po`: one `.po` per language, where `msgstr` holds actual translations.

### Workflow

1. **Get** your source strings in CSV.
2. `pot` &rarr; **generate** `.pot` template.
3. `pos` &rarr; **create** new `.po` files with blank translations.
4. **Edit** `.po` files using Poedit, Crowdin, LLM-assisted scripts,
  or manually.
5. `make` &rarr; **export** translations back into CSV.
6. When new strings appear: `update` &rarr; **merge** into existing `.po`.

Using Git, translators can work on just their own `.po` files in separate
branches/PRs, avoiding noise in other locales.

## Installation

Download the latest binary from releases or build from source:

|           | MacOS                           | Linux                          | Windows                          |
| --------- | ------------------------------- | ------------------------------ | -------------------------------- |
| **AMD64** | [dayz-stringtable-darwin-amd64] | [dayz-stringtable-linux-amd64] | [dayz-stringtable-windows-amd64] |
| **ARM64** | [dayz-stringtable-darwin-arm64] | [dayz-stringtable-linux-arm64] | [dayz-stringtable-windows-arm64] |

```bash
git clone https://github.com/woozymasta/dayz-stringtable.git
cd dayz-stringtable
go build ./cmd/dayz-stringtable
```

Or use Makefile for building:

```bash
make build    # Build for current platform
make release  # Build for all supported platforms
make test     # Run tests
make lint     # Run linter
```

## Usage

```bash
dayz-stringtable [OPTIONS] <command>
```

### Options

* `-v, --version` show version and build info
* `-h, --help` show help

### Commands

#### `pot`

Generate a POT template (empty translations):

```bash
dayz-stringtable pot -i stringtable.csv -o stringtable.pot
```

#### `pos`

Create PO files per language (empty `msgstr`):

```bash
dayz-stringtable pos -i stringtable.csv -o l18n
# or for specific langs:
dayz-stringtable pos -i stringtable.csv -l russian,spanish -o l18n
```

#### `make`

Merge all PO files into one CSV of translations:

```bash
dayz-stringtable make -i stringtable.csv -d l18n -o translated.csv
```

#### `update`

Add new strings from CSV to existing PO files:

```bash
dayz-stringtable update -i stringtable.csv -d l18n
# to a separate folder:
dayz-stringtable update -i stringtable.csv -d l18n -o updated_l18n
```

#### `stats`

Show translation statistics for PO files:

```bash
# Basic statistics for all languages
dayz-stringtable stats -i stringtable.csv -d l18n

# Filter by specific language
dayz-stringtable stats -i stringtable.csv -d l18n -l russian -l english

# Verbose mode with untranslated strings
dayz-stringtable stats -i stringtable.csv -d l18n -V

# JSON format (useful for AI agents and automation)
dayz-stringtable stats -i stringtable.csv -d l18n -cV -f json
```

The `stats` command displays:

* **Translated count**: Number of translated strings
* **Total count**: Total number of strings
* **Completion percentage**: Percentage of translated strings
* **Remaining count**: Number of untranslated strings
* **Untranslated details** (with `--verbose`): List of untranslated strings
  with row numbers, keys, and original text

JSON output format includes all statistics in a structured format suitable
for AI agents and automation scripts.

When used with `--clear-only` flag (`-c`), the command excludes entries
with `# notranslate` comment from translated count, making it useful
for machine translation workflows where you need to identify strings
that actually need translation (excluding intentionally untranslated ones).

#### `clean`

Remove `msgstr` that duplicate `msgid` in PO files:

```bash
dayz-stringtable clean -d l18n
# Only specific languages:
dayz-stringtable clean -d l18n -l russian -l english
# Remove unused entries (not present in CSV):
dayz-stringtable clean -d l18n -i stringtable.csv --remove-unused
# Clear only, don't add notranslate comment:
dayz-stringtable clean -d l18n --clear-only
```

The command scans all `.po` files in the directory
(optionally filtered by `-l`)
and clears `msgstr` where it is identical to `msgid`.
By default, it also adds a `# notranslate` comment to cleaned entries.
Use `--clear-only` to skip adding the comment.
Use `--remove-unused` with `-i`
to remove entries that are no longer present in the CSV file.

#### `translate`

Machine-translate untranslated entries in PO files:

```bash
# DeepL (paid or free API)
dayz-stringtable translate -d l18n deepl --auth-key $DEEPL_AUTH_KEY
dayz-stringtable translate -d l18n deepl --auth-key $DEEPL_AUTH_KEY --api-free

# OpenAI-compatible
dayz-stringtable translate -d l18n openai --api-key $OPENAI_API_KEY

# Google Translate
dayz-stringtable translate -d l18n google --api-key $GOOGLE_TRANSLATE_API_KEY
```

Use `--lang` to target specific languages, `--exclude-lang` to skip originals,
and `--dry-run` to preview counts without calling the provider.

## Integrations & Tools

For integration into your project or CI, you can check out the examples
of automation scripts
[strings.sh](tools/strings.sh) and [strings.ps1](tools/strings.ps1)

* **Poedit** – GUI editor that can auto-translate using DeepL/Google APIs.
* **Crowdin/Lokalise/POEditor** – cloud localization platform.
* **LLMs** – you can script `translate-toolkit` with LibreTranslate or
  OpenAI to auto-fill `msgstr`.

<!-- omit in toc -->
## 👉 [Support Me](https://gist.github.com/WoozyMasta/7b0cabb538236b7307002c1fbc2d94ea)

Your support is greatly appreciated!

<!-- Links -->
[dayz-stringtable-darwin-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-darwin-arm64 "MacOS arm64 file"
[dayz-stringtable-darwin-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-darwin-amd64 "MacOS amd64 file"
[dayz-stringtable-linux-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-linux-amd64 "Linux amd64 file"
[dayz-stringtable-linux-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-linux-arm64 "Linux arm64 file"
[dayz-stringtable-windows-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-windows-amd64.exe "Windows amd64 file"
[dayz-stringtable-windows-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-windows-arm64.exe "Windows arm64 file"
[gettext]: https://www.gnu.org/software/gettext/
