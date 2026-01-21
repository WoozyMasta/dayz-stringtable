# PowerShell script for managing DayZ stringtable translations
param(
    [string]$CSV_TEMPLATE = "",
    [string]$CSV_RESULT = ""
)

$ErrorActionPreference = "Stop"

# Check if dayz-stringtable command exists
try {
    $null = dayz-stringtable -v 2>&1
} catch {
    Write-Error "Not found tool dayz-stringtable"
    Write-Error "https://github.com/WoozyMasta/dayz-stringtable"
    exit 1
}

# Set default variables
if (-not $env:PO_DIR) {
    $env:PO_DIR = "./l18n"
}

if (-not $CSV_TEMPLATE) {
    if ($args.Count -gt 0) {
        $CSV_TEMPLATE = $args[0]
    } else {
        $CSV_TEMPLATE = "./l18n/stringtable.csv"
    }
}

if (-not $CSV_RESULT) {
    if ($args.Count -gt 1) {
        $CSV_RESULT = $args[1]
    } else {
        $CSV_RESULT = "./client/stringtable.csv"
    }
}

$POT_FILE = "./l18n/stringtable.pot"

# Uncomment language for add
$langs = @(
    # "english"
    "czech"
    "german"
    "russian"
    # "polish"
    # "hungarian"
    # "italian"
    # "spanish"
    # "french"
    # "chinese"
    # "japanese"
    # "portuguese"
    # "chinesesimp"
)

# Create CSV template if it doesn't exist
if (-not (Test-Path $CSV_TEMPLATE)) {
    $templateDir = Split-Path -Parent $CSV_TEMPLATE
    if ($templateDir -and -not (Test-Path $templateDir)) {
        New-Item -ItemType Directory -Path $templateDir -Force | Out-Null
    }
    $templateContent = '"Language","original",' + "`n" + '"STR_Yes","Yes",' + "`n" + '"STR_No","No",'
    Set-Content -Path $CSV_TEMPLATE -Value $templateContent
    Write-Host "Init base template $CSV_TEMPLATE"
}

# Join languages with comma
$langsString = $langs -join ","

# Update or create PO files
if ((Test-Path $env:PO_DIR) -and (Test-Path $POT_FILE)) {
    # Update with new strings
    dayz-stringtable update -i $CSV_TEMPLATE -d $env:PO_DIR -l $langsString
} else {
    # First run, create po files
    dayz-stringtable pos -i $CSV_TEMPLATE -d $env:PO_DIR -f -l $langsString
}

dayz-stringtable pot -i $CSV_TEMPLATE -o $POT_FILE -f
dayz-stringtable make -i $CSV_TEMPLATE -d $env:PO_DIR -o $CSV_RESULT -f
dayz-stringtable clean -d $env:PO_DIR
dayz-stringtable stats -i $CSV_TEMPLATE -d $env:PO_DIR
