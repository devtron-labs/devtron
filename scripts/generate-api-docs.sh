#!/bin/bash

# Script to generate HTML documentation from all API specs using Redocly
# Preserves folder structure and generates a card-based index page

set -ex
set -o pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Directories
SPECS_DIR="specs"
OUTPUT_DIR="docs/api-docs"
INDEX_FILE="$OUTPUT_DIR/index.html"
ERROR_LOG="$OUTPUT_DIR/errors.log"

echo -e "${BLUE}🚀 Starting API documentation generation...${NC}"

# Clean output folder
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Check Redocly
if ! command -v redocly &>/dev/null; then
    echo -e "${RED}❌ Redocly CLI not found. Install it:${NC}"
    echo "npm install -g @redocly/cli"
    exit 1
fi
echo -e "${GREEN}✅ Redocly found: $(redocly --version)${NC}"

success_count=0
error_count=0
> "$ERROR_LOG"

convert_spec_to_html() {
    local spec_file="$1"
    local relative_path="${spec_file#$SPECS_DIR/}"
    local output_file="$OUTPUT_DIR/${relative_path%.*}.html"
    mkdir -p "$(dirname "$output_file")"

    echo -e "${BLUE}📄 Converting: $spec_file${NC}"
    if redocly build-docs "$spec_file" -o "$output_file" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Success: $output_file${NC}"
        ((success_count++))
    else
        echo -e "${RED}❌ Failed: $spec_file${NC}"
        echo "$spec_file" >> "$ERROR_LOG"
        ((error_count++))
    fi
}

# Find spec files
mapfile -t spec_files < <(find "$SPECS_DIR" -type f \( -name "*.yaml" -o -name "*.yml" \) | sort)
echo -e "${BLUE}📊 Found ${#spec_files[@]} specs${NC}"

# Convert specs
for spec_file in "${spec_files[@]}"; do
    convert_spec_to_html "$spec_file" || true
done

# Generate index.html
cat > "$INDEX_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Devtron API Documentation</title>
<style>
body { font-family: Arial, sans-serif; margin: 20px; background: #f8f9fa; color: #333; }
h1 { text-align: center; color: #2c3e50; }
h2 { text-align: center; margin-top: 40px; color: #34495e; }
.container { max-width: 1200px; margin: auto; }
.grid { display: flex; flex-wrap: wrap; justify-content: center; gap: 20px; margin-top: 20px; }
.card { background: #fff; border-radius: 8px; padding: 15px; width: calc(25% - 20px); box-shadow: 0 2px 6px rgba(0,0,0,0.1); text-align: center; }
.card a { text-decoration: none; color: #1a73e8; font-weight: bold; }
.card a:hover { text-decoration: underline; }
.footer { margin-top: 40px; font-size: 0.9rem; color: #666; text-align: center; }
.footer a { color: #1a73e8; text-decoration: none; }
.footer a:hover { text-decoration: underline; }
.timestamp { font-style: italic; }
@media(max-width: 1024px){ .card { width: calc(33.33% - 20px); } }
@media(max-width: 768px){ .card { width: calc(50% - 20px); } }
@media(max-width: 480px){ .card { width: 100%; } }
</style>
</head>
<body>
<div class="container">
<h1>🚀 Devtron API Documentation</h1>
<div id="categories"></div>
<div class="footer">
<p><a href="https://devtron.ai/" target="_blank">Devtron</a></p>
<p class="timestamp">Last updated: <span id="timestamp"></span></p>
</div>
</div>
<script>
const apiData = {
EOF

# Populate apiData preserving folder structure
for spec_file in "${spec_files[@]}"; do
    relative_path="${spec_file#$SPECS_DIR/}"
    html_file="${relative_path%.*}.html"
    category=$(dirname "$relative_path")
    [[ "$category" == "." ]] && category="Root"

    display_category=$(echo "$category" | sed 's/[-_]/ /g' | sed 's/\([a-z]\)\([A-Z]\)/\1 \2/g' | sed 's/\b\w/\U&/g')
    title=$(grep -m 1 '^[[:space:]]*title:' "$spec_file" | sed 's/^[[:space:]]*title:[[:space:]]*//' | tr -d '"' || echo "${relative_path%.*}")

    if [[ -f "$OUTPUT_DIR/$html_file" ]]; then
        echo "  \"${category}_$(basename "${relative_path%.*}")\": {\"category\": \"${display_category}\", \"title\": \"${title}\", \"filename\": \"${html_file}\"}," >> "$INDEX_FILE"
    fi
done

sed -i '$ s/,$//' "$INDEX_FILE"

cat >> "$INDEX_FILE" << 'EOF'
};

function populatePage() {
    const container = document.getElementById('categories');
    const categories = {};

    Object.values(apiData).forEach(api => {
        if (!categories[api.category]) categories[api.category] = [];
        categories[api.category].push(api);
    });

    Object.keys(categories).sort().forEach(cat => {
        const heading = document.createElement('h2');
        heading.textContent = cat;
        container.appendChild(heading);

        const grid = document.createElement('div');
        grid.className = "grid";

        categories[cat].sort((a,b)=>a.title.localeCompare(b.title)).forEach(api => {
            const card = document.createElement('div');
            card.className = "card";

            const a = document.createElement('a');
            a.href = api.filename;
            a.textContent = api.title;

            card.appendChild(a);
            grid.appendChild(card);
        });

        container.appendChild(grid);
    });

    document.getElementById('timestamp').textContent = new Date().toLocaleString();
}

document.addEventListener('DOMContentLoaded', populatePage);
</script>
</body>
</html>
EOF

echo -e "${GREEN}✅ Card-based index page generated: $INDEX_FILE${NC}"

# === SUMMARY ===
echo -e "${BLUE}📊 Final Summary:${NC}"
echo -e "${GREEN}✅ Successfully converted: $success_count specs${NC}"
if (( error_count > 0 )); then
    echo -e "${RED}❌ Failed: $error_count (see $ERROR_LOG)${NC}"
fi
echo -e "${BLUE}📁 Output directory: $OUTPUT_DIR${NC}"
echo -e "${BLUE}🌐 Main index: $INDEX_FILE${NC}"

# === CREATE README ===
cat > "$OUTPUT_DIR/README.md" << 'EOF'
# Devtron API Documentation
This folder contains the HTML documentation generated from the OpenAPI specs in the `specs` directory.
EOF

echo -e "${GREEN}✅ README created: $OUTPUT_DIR/README.md${NC}"
echo -e "${GREEN}🎉 API documentation generation complete!${NC}"
