#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null &&  pwd )

cd "${SCRIPT_DIR}"

function goi18n() {
    go run github.com/nicksnyder/go-i18n/v2/goi18n@main "$@"
}

# extract message from Go files into active.en.yaml
# it looks for literal &i18n.Message
goi18n extract -format yaml -sourceLanguage en ./../..

# list of all languages except the default one
EXTRA_LANGS=( "fr" )
for LANG in "${EXTRA_LANGS[@]}"; do
  TRANSLATE_FILE="./translate.${LANG}.yaml"
  ACTIVE_FILE="./active.${LANG}.yaml"
  touch "${TRANSLATE_FILE}"
  # merge values from translate.fr.yaml into active.fr.yaml
  goi18n merge -format yaml -sourceLanguage en active.en.yaml "${ACTIVE_FILE}" "${TRANSLATE_FILE}"
  # create new translate.fr.yaml with missing values
  echo -n "" > "${TRANSLATE_FILE}"
  goi18n merge -format yaml -sourceLanguage en active.en.yaml "${ACTIVE_FILE}" "${TRANSLATE_FILE}"
  # without this, re-running this script would merge the translations with the values from the original lang
  sed -i.bak -r 's#  other:(.+)#  original_other:\1\n  other: ~#' "${TRANSLATE_FILE}"
  rm -f "${TRANSLATE_FILE}.bak"
done
