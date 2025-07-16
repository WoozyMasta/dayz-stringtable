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

|       |    **Linux**    |    **Windows**    |    **macOS**     |
|------:|:---------------:|:-----------------:|:----------------:|
| i386  | [linux-i386][]  | [windows-i386][]  |                  |
| amd64 | [linux-amd64][] | [windows-amd64][] | [darwin-amd64][] |
| arm   | [linux-arm][]   |                   |                  |
| arm64 | [linux-arm64][] | [windows-arm64][] | [darwin-arm64][] |

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

For integration into your project or CI, you can check out the examples
of automation scripts
[strings.sh](tools/strings.sh) and [strings.ps1](tools/strings.ps1)

* **Poedit** – GUI editor that can auto-translate using DeepL/Google APIs.
* **Crowdin/Lokalise/POEditor** – cloud localization platform.
* **LLMs** – you can script `translate-toolkit` with LibreTranslate or
  OpenAI to auto-fill `msgstr`.

<!-- omit in toc -->
### Crypto Donations

<!-- cSpell:disable -->
* **BTC**: `1Jb6vZAMVLQ9wwkyZfx2XgL5cjPfJ8UU3c`
* **USDT (TRC20)**: `TN99xawQTZKraRyvPAwMT4UfoS57hdH8Kz`
* **TON**: `UQBB5D7cL5EW3rHM_44rur9RDMz_fvg222R4dFiCAzBO_ptH`
<!-- cSpell:enable -->

Your support is greatly appreciated!

<!-- Links -->
[darwin-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/darwin-arm64 "MacOS arm64 file"
[darwin-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/darwin-amd64 "MacOS amd64 file"
[linux-i386]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/linux-386 "Linux i386 file"
[linux-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/linux-amd64 "Linux amd64 file"
[linux-arm]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/linux-arm "Linux arm file"
[linux-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/linux-arm64 "Linux arm64 file"
[windows-i386]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/windows-386.exe "Windows i386 file"
[windows-amd64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/windows-amd64.exe "Windows amd64 file"
[windows-arm64]: https://github.com/WoozyMasta/dayz-stringtable/releases/latest/download/windows-arm64.exe "Windows arm64 file"
[gettext]: https://www.gnu.org/software/gettext/
