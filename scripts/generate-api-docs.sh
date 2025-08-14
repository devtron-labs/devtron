```bash
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
/* General body and container styles */
body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    margin: 0;
    padding: 0;
    background-color: #f0f2f5;
    color: #2c3e50;
}
.container {
    max-width: 1200px;
    margin: 20px auto;
    padding: 0 20px;
}
/* Header styles */
.header {
    background-color: #ffffff;
    padding: 20px;
    border-bottom: 1px solid #dfe3e8;
    text-align: center;
    box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}
.header-title {
    font-size: 2.5rem;
    font-weight: 600;
    color: #3b5998;
    margin: 0;
    display: flex;
    align-items: center;
    justify-content: center;
}
.header-title img {
    height: 40px;
    margin-right: 10px;
}
.header-subtitle {
    font-size: 1rem;
    color: #606770;
    margin-top: 10px;
}
/* Grid and card styles */
.grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 20px;
    margin-top: 20px;
}
.card {
    background-color: #ffffff;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    overflow: hidden;
    padding: 20px;
}
.card-header {
    font-size: 1.25rem;
    font-weight: 600;
    color: #3b5998;
    padding-bottom: 10px;
    margin-bottom: 15px;
    border-bottom: 2px solid #3b5998;
}
.card-list {
    list-style: none;
    padding: 0;
    margin: 0;
}
.card-list li {
    margin-bottom: 10px;
}
.card-list a {
    text-decoration: none;
    color: #1877f2;
    font-size: 1rem;
    transition: color 0.2s ease-in-out;
}
.card-list a:hover {
    color: #3b5998;
    text-decoration: underline;
}
/* Footer styles */
.footer {
    margin-top: 40px;
    padding: 20px 0;
    text-align: center;
    border-top: 1px solid #dfe3e8;
    color: #606770;
}
.footer a {
    color: #1877f2;
    text-decoration: none;
}
.footer a:hover {
    text-decoration: underline;
}
.timestamp {
    font-style: italic;
    font-size: 0.9rem;
    color: #8d949e;
}
</style>
</head>
<body>
<div class="header">
    <h1 class="header-title">
        <img src="https://devtron.ai/assets/icons/logo-full.svg" alt="Devtron Logo">
        Devtron API Documentation
    </h1>
    <p class="header-subtitle">Comprehensive API documentation for Devtron - Kubernetes-native software delivery platform</p>
</div>
<div class="container">
<div class="grid" id="categories"></div>
<div class="footer">
<p>
    <a href="https://devtron.ai/" target="_blank">Devtron</a> |
    <a href="https://docs.devtron.ai/" target="_blank">Documentation</a> |
    <a href="https://github.com/devtron-labs/devtron" target="_blank">GitHub</a>
</p>
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
        const card = document.createElement('div');
        card.className = "card";

        const cardHeader = document.createElement('div');
        cardHeader.className = "card-header";
        cardHeader.textContent = cat;
        card.appendChild(cardHeader);

        const list = document.createElement('ul');
        list.className = "card-list";

        categories[cat].sort((a, b) => a.title.localeCompare(b.title)).forEach(api => {
            const listItem = document.createElement('li');
            const a = document.createElement('a');
            a.href = api.filename;
            a.textContent = api.title;
            listItem.appendChild(a);
            list.appendChild(listItem);
        });

        card.appendChild(list);
        container.appendChild(card);
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
```