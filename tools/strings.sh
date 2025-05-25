#!/usr/bin/env bash
set -eu

command dayz-stringtable -v &>/dev/null || {
  >&2 echo "Not found tool dayz-stringtable"
  >&2 echo "https://github.com/WoozyMasta/dayz-stringtable"
  exit 1
}

: "${PO_DIR:=./l18n}"
: "${CSV_TEMPLATE:=${1:-./l18n/stringtable.csv}}"
: "${CSV_RESULT:=${2:-./client/stringtable.csv}}"
: "${POT_FILE:=./l18n/stringtable.pot}"

# uncomment language for add
langs=(
  # czech
  # german
  russian
  # polish
  # hungarian
  # italian
  # spanish
  # french
  # chinese
  # japanese
  # portuguese
  # chinesesimp
)

if [ ! -f "$CSV_TEMPLATE" ]; then
  mkdir -p "$PO_DIR"
  printf '"%s","%s",\n' Language original STR_Yes Yes STR_No No \
    > "$CSV_TEMPLATE"
  echo "Init base template $CSV_TEMPLATE"
fi

if [ -d "$PO_DIR" ] && [ -f "$POT_FILE" ]; then
  # update with new strings
  dayz-stringtable update -i "$CSV_TEMPLATE" -d "$PO_DIR" \
    -l "$( IFS=,; echo "${langs[*]}" )"
else
  # first run, create po files
  dayz-stringtable pos -i "$CSV_TEMPLATE" -d "$PO_DIR" -f \
    -l "$( IFS=,; echo "${langs[*]}" )"
fi

dayz-stringtable pot -i "$CSV_TEMPLATE" -o "$POT_FILE" -f
dayz-stringtable make -i "$CSV_TEMPLATE" -d "$PO_DIR" -o "$CSV_RESULT" -f
