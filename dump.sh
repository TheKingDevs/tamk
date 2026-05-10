#!/data/data/com.termux/files/usr/bin/bash

OUTPUT="dump.md"

# Limpa arquivo anterior
> "$OUTPUT"

echo "# Project Dump" >> "$OUTPUT"
echo "" >> "$OUTPUT"

# Função para checar se é texto
is_text_file() {
    file "$1" | grep -qE 'text|empty'
}

# Diretórios ignorados
IGNORE_DIRS=(
    ".git"
    "__pycache__"
    "node_modules"
    ".venv"
    "venv"
    "build"
    "dist"
)

should_ignore() {
    local path="$1"
    for dir in "${IGNORE_DIRS[@]}"; do
        [[ "$path" == *"/$dir/"* ]] && return 0
    done
    return 1
}

# Percorre arquivos
find . -type f | while read -r file; do
    if should_ignore "$file"; then
        continue
    fi

    if is_text_file "$file"; then
        echo "## FILE: $file" >> "$OUTPUT"
        echo '```' >> "$OUTPUT"
        cat "$file" >> "$OUTPUT"
        echo '' >> "$OUTPUT"
        echo '```' >> "$OUTPUT"
        echo "" >> "$OUTPUT"
    fi
done

echo "Dump criado em $OUTPUT"
