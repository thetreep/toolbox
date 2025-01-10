#!/usr/bin/env bash

set -euo pipefail

goi18n() {
    go run github.com/nicksnyder/go-i18n/v2/goi18n@main "$@"
}

extract_message() {
    local path=$1
    local module_dir
    module_dir=$(go list -m -f '{{.Dir}}')
    # extract message from Go files into active.en.yaml
    # it looks for literal &i18n.Message
    echo "extract in ${path}"
    goi18n extract -format yaml -sourceLanguage en -outdir "${path}" "${module_dir}"
}


# Function to create a file if not exist
create_file() {
    local filepath=$1
    if [ ! -f "$filepath" ]; then
        touch "$filepath"
    fi
}


translate () {
    # list of all languages except the default one
    local extra_langs=$1
    local path=$2
    for lang in "${extra_langs[@]}"; do
      translate_file="${path}translate.${lang}.yaml"
      active_file="${path}active.${lang}.yaml"
      touch "${translate_file}"
      create_file "${active_file}"
      # merge values from translate.fr.yaml into active.fr.yaml
      goi18n merge -format yaml -outdir "${path}" -sourceLanguage en "${path}active.en.yaml" "${active_file}" "${translate_file}"
      # create new translate.fr.yaml with missing values
      echo -n "" > "${translate_file}"
      goi18n merge -format yaml -outdir "${path}" -sourceLanguage en "${path}active.en.yaml" "${active_file}" "${translate_file}"
      # without this, re-running this script would merge the translations with the values from the original lang
      sed -i.bak -r 's#  other:(.+)#  original_other:\1\n  other: ~#' "${translate_file}"
      rm -f "${translate_file}.bak"
    done

}

# Function to display usage information
usage() {
    echo "Usage: $0 <extra languages...> <path>"
    echo "Example: $0 fr en es /path/to/directory/"
    echo "Extra languages: fr, es, ect..."
    exit 1
}

# Function to validate path
validate_path() {
    local path=$1
    # Check if path is provided
    if [ -z "$path" ]; then
        echo "Error: No path provided"
        usage
    fi
    # Check if path exists
    if [ ! -d "$path" ]; then
        echo "Error: Directory '$path' does not exist"
        usage
    fi  
    # Check if path is readable
    if [ ! -r "$path" ]; then
        echo "Error: Directory '$path' is not readable"
        usage
    fi
    # Check if path end with /
    if [[ "${path: -1}" != "/" ]]; then
        echo "Error: Directory '$path' argument doest end with /"
        usage
    fi

    return 0
}

# Validate languages
validate_languages() {
    local provided_langs=("$@")
    # Check if at least one language is provided
    if [ ${#provided_langs[@]} -eq 0 ]; then
        echo "Error: No languages provided"
        usage
    fi

    return 0
}

main() {
    # Check minimum number of arguments (at least 2: one language and path)
    if [ $# -lt 2 ]; then
        echo "Error: Insufficient arguments"
        usage
    fi

    # Get the last argument as path
    local path="${!#}"

    # Get all arguments except the last one as languages
    local languages=("${@:1:$#-1}")

    # Validate inputs
    validate_languages "${languages[@]}"
    validate_path "$path"

    extract_message "${path}"

    translate "${languages[@]}" "${path}"

    echo "done"
}

main "$@"