#!/usr/bin/env pwsh
# stop on errors
$ErrorActionPreference = 'Stop'

if (-not (Get-Command dayz-stringtable -ErrorAction SilentlyContinue)) {
  Write-Error "Not found tool dayz-stringtable"
  Write-Error "https://github.com/WoozyMasta/dayz-stringtable"
  exit 1
}

param(
  [string]$PO_DIR = './l18n',
  [string]$CSV_TEMPLATE = './l18n/stringtable.csv',
  [string]$CSV_RESULT = './client/stringtable.csv',
  [string]$POT_FILE = './l18n/stringtable.pot'
)

# uncomment language for add
$langs = @(
  # 'czech'
  # 'german'
  'russian'
  # 'polish'
  # 'hungarian'
  # 'italian'
  # 'spanish'
  # 'french'
  # 'chinese'
  # 'japanese'
  # 'portuguese'
  # 'chinesesimp'
)

if (-not (Test-Path $CSV_TEMPLATE)) {
  New-Item -ItemType Directory -Force -Path (Split-Path $CSV_TEMPLATE)
  @"
"Language","original",
"STR_Yes","Yes",
"STR_No","No",
"@ | Set-Content -Encoding UTF8 $CSV_TEMPLATE
  Write-Host "Init base template $CSV_TEMPLATE"
}

$langList = $langs -join ','

if (Test-Path $PO_DIR -and (Test-Path $POT_FILE)) {
  # update with new strings
  dayz-stringtable update -i $CSV_TEMPLATE -d $PO_DIR -l $langList
}
else {
  # first run, create po files
  dayz-stringtable pos -i $CSV_TEMPLATE -d $PO_DIR -l $langList -f
}

# всегда обновляем шаблон и собираем итоговый CSV
dayz-stringtable pot -i $CSV_TEMPLATE -o $POT_FILE -f
dayz-stringtable make -i $CSV_TEMPLATE -d $PO_DIR -o $CSV_RESULT -f
