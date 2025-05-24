# DayZ StringTable GetText CLI

A simple CLI tool for managing localization of DayZ `stringtables.csv`
and Gettext formats.

This utility helps you:

* **Generate a POT template** from your base CSV of original strings.
* **Create PO files** (per language) from CSV, with empty translations.
* **Merge PO files** back into a full CSV of translations.
* **Update existing PO** files when you add new strings to CSV.

Under the hood it uses [gettext] `.pot` and `.po` formats.

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

|         **Download links**         |
| :--------------------------------: |
| [dayz-stringtable-darwin-arm64][]  |
| [dayz-stringtable-darwin-amd64][]  |
|  [dayz-stringtable-linux-i386][]   |
|  [dayz-stringtable-linux-amd64][]  |
|   [dayz-stringtable-linux-arm][]   |
|  [dayz-stringtable-linux-arm64][]  |
| [dayz-stringtable-windows-i386][]  |
| [dayz-stringtable-windows-amd64][] |
| [dayz-stringtable-windows-arm64][] |

```bash
git clone https://github.com/woozymasta/dayz-stringtable.git
cd dayz-stringtable
go build ./cmd/dayz-stringtable
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

## Integrations & Tools

* **Poedit** – GUI editor that can auto-translate using DeepL/Google APIs.
* **Crowdin/Lokalise/POEditor** – cloud localization platform.
* **LLMs** – you can script `translate-toolkit` with LibreTranslate or
  OpenAI to auto-fill `msgstr`.

<!-- Links -->
[dayz-stringtable-darwin-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-darwin-arm64 "MacOS arm64 file"
[dayz-stringtable-darwin-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-darwin-amd64 "MacOS amd64 file"
[dayz-stringtable-linux-i386]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-linux-386 "Linux i386 file"
[dayz-stringtable-linux-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-linux-amd64 "Linux amd64 file"
[dayz-stringtable-linux-arm]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-linux-arm "Linux arm file"
[dayz-stringtable-linux-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-linux-arm64 "Linux arm64 file"
[dayz-stringtable-windows-i386]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-windows-386.exe "Windows i386 file"
[dayz-stringtable-windows-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-windows-amd64.exe "Windows amd64 file"
[dayz-stringtable-windows-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/dayz-stringtable-windows-arm64.exe "Windows arm64 file"
[gettext]: https://www.gnu.org/software/gettext/
