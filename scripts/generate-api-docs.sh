#!/bin/bash

# Script to generate HTML documentation from all API specs using Redocly
# This will convert all .yaml and .yml files in the specs directory to HTML

set -ex
set -o pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
SPECS_DIR="specs"
OUTPUT_DIR="docs/api-docs"
INDEX_FILE="$OUTPUT_DIR/index.html"
ERROR_LOG="$OUTPUT_DIR/errors.log"

echo -e "${BLUE}🚀 Starting API documentation generation...${NC}"

# === CLEAN OUTPUT DIRECTORY ===
echo -e "${YELLOW}🧹 Cleaning output folder...${NC}"
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Check if redocly is installed
if ! command -v redocly &> /dev/null; then
    echo -e "${RED}❌ Redocly is not installed. Please install it first:${NC}"
    echo "npm install -g @redocly/cli"
    exit 1
fi
echo -e "${GREEN}✅ Redocly found: $(redocly --version)${NC}"

# Counters
success_count=0
error_count=0

# Clear error log
> "$ERROR_LOG"

# Function to convert a spec file to HTML
convert_spec_to_html() {
    local spec_file="$1"
    local relative_path="${spec_file#$SPECS_DIR/}"
    local filename=$(basename "$spec_file")
    local name_without_ext="${filename%.*}"
    local output_file="$OUTPUT_DIR/${relative_path%.*}.html"

    # Create output directory if it doesn't exist
    mkdir -p "$(dirname "$output_file")"

    echo -e "${BLUE}📄 Converting: $spec_file${NC}"

    if redocly build-docs "$spec_file" -o "$output_file" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Success: $output_file${NC}"
        ((success_count++))
        return 0
    else
        echo -e "${RED}❌ Failed: $spec_file${NC}"
        echo "$spec_file" >> "$ERROR_LOG"
        ((error_count++))
        return 1
    fi
}

# === FIND AND CONVERT SPEC FILES ===
echo -e "${YELLOW}🔍 Finding all spec files...${NC}"
mapfile -t spec_files < <(find "$SPECS_DIR" -type f \( -name "*.yaml" -o -name "*.yml" \) | sort)
echo -e "${BLUE}📊 Found ${#spec_files[@]} spec files${NC}"

for spec_file in "${spec_files[@]}"; do
    convert_spec_to_html "$spec_file" || true
done

# === GENERATE INDEX.HTML ===
echo -e "${YELLOW}📝 Generating index page...${NC}"

cat > "$INDEX_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Devtron API Documentation</title>
<style>
body { font-family: Arial, sans-serif; margin: 20px; }
h1 { color: #333; }
h3 { margin-top: 20px; }
ul { list-style: none; padding-left: 0; }
li { margin: 5px 0; }
a { text-decoration: none; color: #0366d6; }
a:hover { text-decoration: underline; }
</style>
</head>
<body>
<div class="container">
<h1>🚀 Devtron API Documentation</h1>
<div class="categories" id="categories"></div>
<div class="footer">
<p><a href="https://devtron.ai/" target="_blank">Devtron</a></p>
<p class="timestamp">Last updated: <span id="timestamp"></span></p>
</div>
</div>
<script>
const apiData = {
EOF

# Populate apiData
for spec_file in "${spec_files[@]}"; do
    relative_path="${spec_file#$SPECS_DIR/}"
    html_file="${relative_path%.*}.html"
    category=$(dirname "$relative_path")
    [[ "$category" == "." ]] && category="root"

    # Capitalise each word and split camelCase
    display_category=$(echo "$category" | sed 's/[-_]/ /g' | sed 's/\([a-z]\)\([A-Z]\)/\1 \2/g' | sed 's/\b\w/\U&/g')

    # Get title or fallback
    title=$(grep -m 1 '^[[:space:]]*title:' "$spec_file" | sed 's/^[[:space:]]*title:[[:space:]]*//' | tr -d '"' || echo "${relative_path%.*}")

    if [[ -f "$OUTPUT_DIR/$html_file" ]]; then
        echo "  \"${category}_$(basename "${relative_path%.*}")\": {\"category\": \"${display_category}\", \"title\": \"${title}\", \"filename\": \"${html_file}\"}," >> "$INDEX_FILE"
    fi
done

# Remove trailing comma
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
        const section = document.createElement('div');
        const h3 = document.createElement('h3');
        h3.textContent = cat;
        section.appendChild(h3);

        const ul = document.createElement('ul');
        categories[cat].sort((a,b)=>a.title.localeCompare(b.title)).forEach(api => {
            const li = document.createElement('li');
            const a = document.createElement('a');
            a.href = api.filename;
            a.textContent = api.title;
            li.appendChild(a);
            ul.appendChild(li);
        });

        section.appendChild(ul);
        container.appendChild(section);
    });

    document.getElementById('timestamp').textContent = new Date().toLocaleString();
}

document.addEventListener('DOMContentLoaded', populatePage);
</script>
</body>
</html>
EOF

echo -e "${GREEN}✅ Index page generated: $INDEX_FILE${NC}"

# === FINAL SUMMARY ===
echo -e "${BLUE}📊 Final Summary:${NC}"
echo -e "${GREEN}✅ Successfully converted: $success_count files${NC}"
if [ $error_count -gt 0 ]; then
    echo -e "${RED}❌ Failed to convert: $error_count files${NC}"
    echo -e "${YELLOW}📝 Check $ERROR_LOG for details${NC}"
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
