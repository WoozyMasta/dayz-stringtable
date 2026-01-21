# AI Agent Instructions: DayZ StringTable Localization Tool

## Project Overview

**DayZ StringTable** is a CLI tool designed to manage localization workflows
for DayZ game stringtables. It converts between two formats:

- **CSV format**: The native format used by DayZ (`stringtable.csv`)
- **Gettext PO/POT format**: Standard translation format used by tools like
  Poedit, Crowdin, and translation workflows

### Purpose

This tool enables AI agents and automation scripts to:

- Extract translatable strings from DayZ CSV files
- Work with standard Gettext PO files for translation
- Merge translations back into DayZ-compatible CSV format
- Track translation progress and identify untranslated strings
- Automate translation workflows

### Key Use Cases for Automation

- **Translation Management**: Convert CSV to PO, translate, then merge back to CSV
- **Progress Tracking**: Identify which strings need translation
- **Batch Processing**: Update multiple language files simultaneously
- **CI/CD Integration**: Automate translation checks and updates in pipelines

## Quick Start for AI Agents

### Prerequisites

1. **Install dayz-stringtable binary**:
  The tool must be installed and available in PATH
   - Check installation: `dayz-stringtable -v`
   - If not installed, download from:
     <https://github.com/WoozyMasta/dayz-stringtable/releases>

2. **Required Files**:
   - Source CSV file (e.g., `stringtable.csv`) with structure: `Language,original,<lang1>,<lang2>,...`
   - PO directory (default: `./l18n`) containing `.po` files per language

### Recommended Workflow: Helper Scripts

**IMPORTANT**: Before starting any translation work, and after completing work,
call the appropriate helper script:

- **Windows**: `tools/strings.ps1`
- **Unix/Linux/MacOS**: `tools/strings.sh`

These scripts automate the complete workflow:

1. Update/create PO files from CSV
2. Generate POT template
3. Merge PO files back to CSV
4. Clean duplicate translations
5. Display statistics

**Usage**:

```bash
# Before starting work
./tools/strings.sh [input_csv] [output_csv]

# After completing work
./tools/strings.sh [input_csv] [output_csv]
```

**Script Parameters** (optional):

- First argument: Input CSV template (default: `./l18n/stringtable.csv`)
- Second argument: Output CSV result (default: `./client/stringtable.csv`)

**Environment Variables**:

- `PO_DIR`: PO files directory (default: `./l18n`)

## Core Commands Reference

### `pot` - Generate POT Template

Creates a Gettext template file (`.pot`) from CSV
with all original strings and empty translations.

```bash
dayz-stringtable pot -i stringtable.csv -o stringtable.pot
```

**When to use**: Initial setup, creating translation templates.

### `pos` - Create PO Files Per Language

Generates individual `.po` files for each language from CSV,
with empty `msgstr` fields.

```bash
# Create PO files for all languages
dayz-stringtable pos -i stringtable.csv -d l18n

# Create PO files for specific languages
dayz-stringtable pos -i stringtable.csv -d l18n -l russian,spanish
```

**When to use**: Initial setup, adding new languages.

### `make` - Merge PO Files to CSV

Combines all PO files back into a single CSV file with translations.

```bash
dayz-stringtable make -i stringtable.csv -d l18n -o translated.csv
```

**When to use**:
After translations are complete, to generate final CSV for DayZ.

### `update` - Update Existing PO Files

Adds new strings from CSV to existing PO files
without losing existing translations.

```bash
# Update in-place
dayz-stringtable update -i stringtable.csv -d l18n

# Update to separate directory
dayz-stringtable update -i stringtable.csv -d l18n -o updated_l18n
```

**When to use**: When new strings are added to CSV, to sync PO files.

### `stats` - Show Translation Statistics

Displays translation completion statistics. **Critical for AI agents**
to identify untranslated strings.

```bash
# Basic statistics (table format)
dayz-stringtable stats -i stringtable.csv -d l18n

# Filter by specific languages
dayz-stringtable stats -i stringtable.csv -d l18n -l russian -l english

# Verbose mode with untranslated strings (text format)
dayz-stringtable stats -i stringtable.csv -d l18n -V

# JSON format with untranslated strings (recommended for automation)
dayz-stringtable stats -i stringtable.csv -d l18n -f json -V
```

**When to use**:

- Before translation: Identify what needs translation
- After translation: Verify completion status
- In automation: Parse JSON output for programmatic processing

### `clean` - Remove Duplicate Translations

Removes `msgstr` entries that duplicate `msgid`
(untranslated strings that match originals).

```bash
# Clean all languages
dayz-stringtable clean -d l18n

# Clean specific languages
dayz-stringtable clean -d l18n -l russian -l english

# Remove unused entries (not in CSV)
dayz-stringtable clean -d l18n -i stringtable.csv --remove-unused
```

**When to use**: After translation, to clear false translations
(where translation equals original).

## Getting Untranslated Strings

The `stats` command provides two output formats
for identifying untranslated strings. Both are essential for AI agents.

### Text Format (Grep-style)

**Command**:

```bash
dayz-stringtable stats -i stringtable.csv -d l18n -V
```

**Output Format**: `po_file:line:key:"original_text"`

**Example Output**:

```txt
russian.po:45:STR_Yes:"Yes"
russian.po:67:STR_No:"No"
spanish.po:45:STR_Yes:"Yes"
```

**Parsing**:
Each line follows the pattern
`filename:line_number:translation_key:"original_text"`.
This format is compatible with `grep -nr` style output,
making it easy to process with standard text tools.

**Use Case**: Quick inspection, grep-style filtering, simple text processing.

### JSON Format (Recommended for Automation)

**Command**:

```bash
dayz-stringtable stats -i stringtable.csv -d l18n -f json -V
```

**Output Structure**:

```json
{
  "languages": {
    "russian": {
      "translated": 150,
      "total": 200,
      "percentage": 75.0,
      "remaining": 50,
      "untranslated": [
        {
          "key": "STR_Yes",
          "original": "Yes",
          "context": "STR_Yes",
          "po_file": "russian.po",
          "row": 2,
          "po_line": 45
        },
        {
          "key": "STR_No",
          "original": "No",
          "context": "STR_No",
          "po_file": "russian.po",
          "row": 3,
          "po_line": 67
        }
      ]
    },
    "spanish": {
      "translated": 180,
      "total": 200,
      "percentage": 90.0,
      "remaining": 20,
      "untranslated": [...]
    }
  }
}
```

**Field Descriptions**:

- `translated`: Number of translated strings
- `total`: Total number of strings
- `percentage`: Completion percentage (0-100)
- `remaining`: Number of untranslated strings
- `untranslated`: Array of untranslated items (only present with `-V` flag)
  - `key`: Translation key (CSV first column)
  - `original`: Original text (CSV second column)
  - `context`: Context identifier (same as key)
  - `po_file`: PO filename (e.g., "russian.po")
  - `row`: CSV row number (1-based, including header)
  - `po_line`: Line number in PO file where entry is located

**Use Case**:
Programmatic processing, AI agent automation, structured data analysis.

**Note**:
The `untranslated` array is only included when using the `-V` (verbose) flag.
Without `-V`, only statistics are provided.

### Filtering by Language

Both formats support filtering by specific languages:

```bash
# Text format
dayz-stringtable stats -i stringtable.csv -d l18n -l russian -V

# JSON format
dayz-stringtable stats -i stringtable.csv -d l18n -l russian -f json -V
```

### Clear-Only Mode

When using `--clear-only` flag (`-c`),
entries with `# notranslate` comments are excluded from translated count:

```bash
dayz-stringtable stats -i stringtable.csv -d l18n -f json -V -c
```

This is useful for machine translation workflows where you need to
identify strings that actually need translation,
excluding intentionally untranslated ones.

## Automation Workflow

### Typical AI Agent Translation Workflow

1. **Initialize/Update Environment**

   ```bash
   # Call helper script to sync PO files with CSV
   ./tools/strings.sh stringtable.csv translated.csv
   ```

2. **Get Untranslated Strings**

   ```bash
   # Get list of untranslated strings in JSON format
   dayz-stringtable stats -i stringtable.csv -d l18n -f json -V > untranslated.json
   ```

3. **Process Translations**
   - Parse JSON to identify untranslated strings
   - Translate strings (using AI, API, or manual process)
   - Update PO files directly or use translation tools

4. **Update PO Files**
   - Edit `.po` files in `l18n/` directory
   - Or use automated translation scripts to fill `msgstr` fields

5. **Finalize and Verify**

   ```bash
   # Call helper script to merge and verify
   ./tools/strings.sh stringtable.csv translated.csv
   ```

### Step-by-Step Automation Example

```bash
#!/bin/bash
# Example automation script

CSV_INPUT="./l18n/stringtable.csv"
CSV_OUTPUT="./client/stringtable.csv"
PO_DIR="./l18n"

# Step 1: Sync PO files with latest CSV
./tools/strings.sh "$CSV_INPUT" "$CSV_OUTPUT"

# Step 2: Get untranslated strings for Russian
dayz-stringtable stats -i "$CSV_INPUT" -d "$PO_DIR" -l russian -f json -V > untranslated_ru.json

# Step 3: Process translations (your AI/automation logic here)
# ... translate strings and update russian.po ...

# Step 4: Verify and merge
./tools/strings.sh "$CSV_INPUT" "$CSV_OUTPUT"

# Step 5: Check final statistics
dayz-stringtable stats -i "$CSV_INPUT" -d "$PO_DIR" -f json
```

### Integration with Translation APIs

1. Get untranslated strings in JSON format
2. Extract `original` field from each untranslated item
3. Call translation API (OpenAI, DeepL, Google Translate, etc.)
4. Update corresponding PO file's `msgstr` field
5. Run helper script to merge and verify

## File Formats

### CSV Format Structure

**Header Row**: `"Language","original","<lang1>","<lang2>",...`

**Data Rows**: `"<key>","<original_text>","<translation1>","<translation2>",...`

**Example**:

```csv
"Language","original","russian","spanish"
"STR_Yes","Yes","Да","Sí"
"STR_No","No","Нет","No"
```

**Column Mapping**:

- Column 0: Translation key (e.g., "STR_Yes")
- Column 1: Original text (source language, typically English)
- Column 2+: Translations for each language

### PO File Format Structure

PO files use standard Gettext format:

```gettext
msgctxt "STR_Yes"
msgid "Yes"
msgstr "Да"
```

**Field Mapping**:

- `msgctxt`: Translation key (maps to CSV column 0)
- `msgid`: Original text (maps to CSV column 1)
- `msgstr`: Translation (maps to CSV language column)

**File Naming**: `<language>.po` (e.g., `russian.po`, `spanish.po`)

**Directory Structure**:

```txt
l18n/
  ├── russian.po
  ├── spanish.po
  ├── german.po
  └── ...
```

### Format Conversion Flow

```txt
CSV (DayZ format)
    ↓ [pot/pos commands]
PO files (Gettext format)
    ↓ [translation work]
Updated PO files
    ↓ [make command]
CSV with translations (DayZ format)
```

## Best Practices

### Error Handling

1. **Check tool availability**:

   ```bash
   if ! command -v dayz-stringtable &> /dev/null; then
       echo "Error: dayz-stringtable not found"
       exit 1
   fi
   ```

2. **Validate file existence**:
   - Always check if CSV input file exists
   - Verify PO directory exists before operations
   - Handle missing PO files gracefully

3. **Check command exit codes**:
   - All commands return non-zero exit code on error
   - Check `$?` after each command in scripts

### File Path Conventions

- **Input CSV**:
  Typically `./l18n/stringtable.csv` or `./stringtable.csv`
- **Output CSV**:
  Typically `./client/stringtable.csv` or `./translated.csv`
- **PO Directory**:
  Default `./l18n`, can be overridden with `PO_DIR` environment variable
- **POT Template**:
  Typically `./l18n/stringtable.pot`

### Working with Multiple Languages

1. **List available languages**:

   ```bash
   ls l18n/*.po | sed 's/.*\///;s/\.po$//'
   ```

2. **Process languages individually**:

   ```bash
   for lang in russian spanish german; do
       dayz-stringtable stats -i stringtable.csv -d l18n -l "$lang" -f json -V
   done
   ```

3. **Batch operations**:
   - Use `-l` flag with comma-separated list: `-l russian,spanish,german`
   - Or process all languages by omitting `-l` flag

### Translation Quality

1. **Clean duplicates**:
   Always run `clean` command after translation to remove false translations
2. **Verify statistics**:
   Check completion percentages before finalizing
3. **Preserve context**:
   PO files preserve comments and metadata - be careful when editing
4. **Use `--clear-only` flag**:
   When identifying strings that need actual translation
   (excludes `# notranslate` entries)

### Performance Considerations

- PO files are parsed on each command execution
- For large projects, consider caching statistics
- JSON output is more efficient for programmatic processing than parsing text output
- Helper scripts run multiple commands sequentially -
  consider parallelization for very large projects

## Common Automation Scenarios

### Scenario 1: Initial Setup

```bash
# 1. Create PO files from CSV
dayz-stringtable pos -i stringtable.csv -d l18n -l russian,spanish

# 2. Generate POT template
dayz-stringtable pot -i stringtable.csv -o l18n/stringtable.pot

# 3. Check initial status
dayz-stringtable stats -i stringtable.csv -d l18n -f json
```

### Scenario 2: Adding New Strings

```bash
# 1. Update PO files with new strings from CSV
dayz-stringtable update -i stringtable.csv -d l18n

# 2. Get only new untranslated strings
dayz-stringtable stats -i stringtable.csv -d l18n -f json -V > new_untranslated.json

# 3. Translate new strings (your process)

# 4. Merge back to CSV
dayz-stringtable make -i stringtable.csv -d l18n -o translated.csv
```

### Scenario 3: Translation Progress Monitoring

```bash
# Get comprehensive statistics
dayz-stringtable stats -i stringtable.csv -d l18n -f json > stats.json

# Parse JSON to identify languages needing attention
# Languages with low percentage or high remaining count
```

### Scenario 4: Complete Workflow Automation

```bash
# Use helper script for complete workflow
./tools/strings.sh stringtable.csv translated.csv

# This automatically:
# - Updates/creates PO files
# - Generates POT template
# - Merges translations to CSV
# - Cleans duplicates
# - Shows statistics
```

## Summary for AI Agents

**Key Takeaways**:

1. **Always call helper scripts**
   (`tools/strings.ps1` or `tools/strings.sh`)
   before starting work and after completing work
2. **Use JSON format**
   (`-f json -V`) for programmatic access to untranslated strings
3. **Text format** (`-V`) provides grep-style output for quick inspection
4. **Workflow**: CSV → PO → Translate → PO → CSV
5. **Statistics are essential** for identifying what needs translation
6. **Clean duplicates** after translation to maintain quality

**Essential Commands**:

- `stats -f json -V`: Get untranslated strings (automation)
- `update`: Sync PO files with new CSV strings
- `make`: Merge translations back to CSV
- Helper scripts: Complete workflow automation

**File Locations**:

- Helper scripts: `tools/strings.ps1`, `tools/strings.sh`
- Default PO directory: `./l18n`
- Default input CSV: `./l18n/stringtable.csv` or `./stringtable.csv`

For more details, see the main [README.md](../README.md) file.
