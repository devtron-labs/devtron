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
    echo -e "${BLUE}    → Output: ${relative_path%.*}.html${NC}"

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
        h1 { text-align: center; color: #2c3e50; margin-bottom: 40px; }
        .container { max-width: 1200px; margin: auto; }
        .categories-grid {
            display: flex;
            flex-wrap: wrap;
            justify-content: center;
            gap: 30px;
            margin-top: 20px;
            align-items: stretch; /* Ensures all cards stretch to same height */
        }

        /* Single card centering */
        .categories-grid.single-card {
            justify-content: center;
            max-width: 400px;
            margin: 20px auto;
        }

        /* Category Cards */
        .category-card {
            background: #fff;
            border-radius: 12px;
            padding: 25px;
            width: calc(33.33% - 30px);
            min-width: 300px;
            max-width: 400px;
            height: auto;
            min-height: 300px; /* Minimum height for consistency */
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            border: 1px solid #e1e5e9;
            transition: transform 0.2s ease, box-shadow 0.2s ease;
            display: flex;
            flex-direction: column;
        }
        .category-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(0,0,0,0.15);
        }

        /* Single card styling */
        .category-card.single {
            width: 100%;
            max-width: 400px;
        }

        /* Category Headers */
        .category-header {
            color: #2c3e50;
            font-size: 1.4em;
            font-weight: bold;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 2px solid #3498db;
            text-align: center;
        }

        /* API Links within Categories */
        .api-links {
            display: flex;
            flex-direction: column;
            gap: 8px;
            flex-grow: 1; /* Takes up remaining space in the card */
            overflow-y: auto; /* Allows scrolling if too many links */
            max-height: 400px; /* Maximum height before scrolling */
        }
        .api-link {
            display: block;
            padding: 8px 12px;
            background: #f8f9fa;
            border-radius: 6px;
            text-decoration: none;
            color: #1a73e8;
            font-weight: 500;
            transition: all 0.2s ease;
            border-left: 3px solid transparent;
        }
        .api-link:hover {
            background: #e3f2fd;
            border-left-color: #1a73e8;
            text-decoration: none;
            transform: translateX(5px);
        }

        /* Footer */
        .footer {
            margin-top: 50px;
            font-size: 0.9rem;
            color: #666;
            text-align: center;
            padding-top: 20px;
            border-top: 1px solid #e1e5e9;
        }
        .footer a { color: #1a73e8; text-decoration: none; }
        .footer a:hover { text-decoration: underline; }
        .timestamp { font-style: italic; }

        /* Responsive Design */
        @media(max-width: 1024px){
            .category-card { width: calc(50% - 30px); }
            .categories-grid.single-card { max-width: 500px; }
        }
        @media(max-width: 768px){
            .category-card { width: 100%; min-width: unset; max-width: none; }
            .categories-grid.single-card { max-width: 100%; margin: 20px 0; }
        }
        @media(max-width: 480px){
            .category-card { margin: 0 10px; }
            .categories-grid { gap: 20px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1> Devtron API Documentation</h1>
        <div id="categories" class="categories-grid"></div>
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

    # Only include if HTML file was successfully generated
    if [[ -f "$OUTPUT_DIR/$html_file" ]]; then
        # Ensure proper relative path from index.html to the generated HTML file
        # Since index.html is in docs/api-docs/ and HTML files maintain folder structure
        echo " \"${category}_$(basename "${relative_path%.*}")\": {\"category\": \"${display_category}\", \"title\": \"${title}\", \"filename\": \"${html_file}\"}," >> "$INDEX_FILE"
    fi
done

sed -i '$ s/,$//' "$INDEX_FILE"



cat >> "$INDEX_FILE" << 'EOF'
        };

        function populatePage() {
            const container = document.getElementById('categories');
            const categories = {};

            // Group APIs by category
            Object.values(apiData).forEach(api => {
                if (!categories[api.category]) categories[api.category] = [];
                categories[api.category].push(api);
            });

            const categoryNames = Object.keys(categories).sort();

            // Add class for single card centering
            if (categoryNames.length === 1) {
                container.classList.add('single-card');
            }

            // Create category cards
            categoryNames.forEach(categoryName => {
                // Create category card
                const categoryCard = document.createElement('div');
                categoryCard.className = 'category-card';

                // Add single class if only one card
                if (categoryNames.length === 1) {
                    categoryCard.classList.add('single');
                }

                // Create category header
                const categoryHeader = document.createElement('div');
                categoryHeader.className = 'category-header';
                categoryHeader.textContent = categoryName;
                categoryCard.appendChild(categoryHeader);

                // Create links container
                const linksContainer = document.createElement('div');
                linksContainer.className = 'api-links';

                // Add API links to this category
                categories[categoryName]
                    .sort((a, b) => a.title.localeCompare(b.title))
                    .forEach(api => {
                        const apiLink = document.createElement('a');
                        // Ensure proper relative path
                        apiLink.href = api.filename;
                        apiLink.textContent = api.title;
                        apiLink.className = 'api-link';
                        apiLink.title = `View ${api.title} API documentation`;

                        // Add click handler to check if file exists
                        apiLink.addEventListener('click', function(e) {
                            // Let the browser handle the navigation normally
                            // This is just for debugging - remove in production if needed
                            console.log(`Navigating to: ${api.filename}`);
                        });

                        linksContainer.appendChild(apiLink);
                    });

                categoryCard.appendChild(linksContainer);
                container.appendChild(categoryCard);
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