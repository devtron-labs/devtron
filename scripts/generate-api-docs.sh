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
spec_files=()
while IFS= read -r -d '' file; do
    spec_files+=("$file")
done < <(find "$SPECS_DIR" -type f \( -name "*.yaml" -o -name "*.yml" \) -print0 | sort -z)

# Fallback if the above doesn't work
if [ ${#spec_files[@]} -eq 0 ]; then
    while IFS= read -r file; do
        spec_files+=("$file")
    done < <(find "$SPECS_DIR" -type f \( -name "*.yaml" -o -name "*.yml" \) | sort)
fi

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
        * { box-sizing: border-box; }

        body {
            font-family: 'Inter', 'Roboto', 'Segoe UI', sans-serif;
            margin: 0;
            padding: 0;
            background: #f9f9f9;
            min-height: 100vh;
            color: #333;
            line-height: 1.6;
        }

        h1 {
            text-align: left;
            color: #333;
            margin: 0 0 32px 0;
            font-weight: 600;
            font-size: 2.5rem;
            border-bottom: 3px solid rgb(0, 102, 204);
            padding-bottom: 16px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 32px;
            background: white;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border-radius: 8px;
            margin-top: 20px;
            margin-bottom: 20px;
        }

        .categories-grid {
            display: flex;
            flex-direction: column;
            gap: 16px;
            margin-top: 24px;
        }

        /* Category Sections */
        .category-section {
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            background: white;
            overflow: hidden;
        }

        .category-summary {
            padding: 16px;
            background: #f8f9fa;
            border-bottom: 1px solid #e5e7eb;
            cursor: pointer;
            display: flex;
            align-items: center;
            font-size: 20px;
            font-weight: 600;
            color: #333;
            transition: background-color 0.2s ease;
            user-select: none;
        }

        .category-summary:hover {
            background: #eef5ff;
        }

        .category-icon {
            margin-right: 12px;
            font-size: 18px;
        }

        .category-toggle {
            margin-left: auto;
            font-size: 14px;
            color: #666;
            transition: transform 0.2s ease;
        }

        .category-section[open] .category-toggle {
            transform: rotate(180deg);
        }

        .category-content {
            padding: 0;
        }

        .api-list {
            list-style: none;
            margin: 0;
            padding: 0;
        }

        .api-item {
            border-bottom: 1px solid #f1f3f4;
        }

        .api-item:last-child {
            border-bottom: none;
        }

        .api-link {
            display: block;
            padding: 16px 20px;
            text-decoration: none;
            color: #333;
            font-size: 16px;
            font-weight: 400;
            transition: all 0.2s ease;
            border-left: 3px solid transparent;
            position: relative;
        }

        .api-link:hover {
            background: #eef5ff;
            border-left-color: rgb(0, 102, 204);
            color: #333;
            padding-left: 24px;
        }

        .api-link:active {
            background: #dbeafe;
        }



        /* Footer */
        .footer {
            margin-top: 40px;
            font-size: 0.85rem;
            color: #666;
            text-align: center;
            padding: 20px 0;
            border-top: 1px solid #e5e7eb;
        }

        .footer a {
            color: rgb(0, 102, 204);
            text-decoration: none;
            font-weight: 500;
        }

        .footer a:hover {
            text-decoration: underline;
        }

        .timestamp {
            font-style: italic;
            opacity: 0.7;
        }

        /* Responsive Design */
        @media(max-width: 768px){
            .container {
                padding: 20px;
                margin: 10px;
            }
            h1 {
                font-size: 2rem;
                margin-bottom: 24px;
            }
            .category-summary {
                padding: 12px;
                font-size: 18px;
            }
            .api-link {
                padding: 12px 16px;
                font-size: 15px;
            }
        }

        @media(max-width: 480px){
            .container {
                padding: 16px;
                margin: 5px;
            }
            h1 {
                font-size: 1.75rem;
            }
            .category-summary {
                padding: 10px;
                font-size: 16px;
            }
            .api-link {
                padding: 10px 14px;
                font-size: 14px;
            }
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

# Remove trailing comma from last line
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS sed
    sed -i '' '$ s/,$//' "$INDEX_FILE"
else
    # Linux sed
    sed -i '$ s/,$//' "$INDEX_FILE"
fi



cat >> "$INDEX_FILE" << 'EOF'
        };

        function getCategoryIcon(categoryName) {
            const icons = {
                'application': '📦',
                'audit': '📝',
                'authentication': '🔑',
                'bulk': '📋',
                'charts': '📊',
                'cluster': '🖥️',
                'deployment': '🚀',
                'ent-only': '🏢',
                'environment': '🌍',
                'external-app': '🔗',
                'fluxcd': '🔄',
                'gitops': '🔀',
                'global-config': '⚙️',
                'helm': '⚓',
                'bulkedit': '✏️'
            };
            return icons[categoryName.toLowerCase()] || '📄';
        }

        function populatePage() {
            const container = document.getElementById('categories');
            const categories = {};

            // Group APIs by category
            Object.values(apiData).forEach(api => {
                if (!categories[api.category]) categories[api.category] = [];
                categories[api.category].push(api);
            });

            const categoryNames = Object.keys(categories).sort();

            // Create category sections
            categoryNames.forEach(categoryName => {
                // Create details element for accordion
                const categorySection = document.createElement('details');
                categorySection.className = 'category-section';
                categorySection.open = true; // Open by default

                // Create summary (header)
                const categorySummary = document.createElement('summary');
                categorySummary.className = 'category-summary';

                const categoryIcon = document.createElement('span');
                categoryIcon.className = 'category-icon';
                categoryIcon.textContent = getCategoryIcon(categoryName);

                const categoryTitle = document.createElement('span');
                categoryTitle.textContent = categoryName.charAt(0).toUpperCase() + categoryName.slice(1);

                const toggleIcon = document.createElement('span');
                toggleIcon.className = 'category-toggle';
                toggleIcon.textContent = '▼';

                categorySummary.appendChild(categoryIcon);
                categorySummary.appendChild(categoryTitle);
                categorySummary.appendChild(toggleIcon);

                // Create content container
                const categoryContent = document.createElement('div');
                categoryContent.className = 'category-content';

                // Create API list
                const apiList = document.createElement('ul');
                apiList.className = 'api-list';

                // Sort APIs and add to list
                const sortedApis = categories[categoryName].sort((a, b) => a.title.localeCompare(b.title));

                sortedApis.forEach(api => {
                    const apiItem = document.createElement('li');
                    apiItem.className = 'api-item';

                    const apiLink = document.createElement('a');
                    apiLink.href = api.filename;
                    apiLink.textContent = api.title;
                    apiLink.className = 'api-link';
                    apiLink.title = `View ${api.title} API documentation`;

                    apiItem.appendChild(apiLink);
                    apiList.appendChild(apiItem);
                });

                categoryContent.appendChild(apiList);
                categorySection.appendChild(categorySummary);
                categorySection.appendChild(categoryContent);
                container.appendChild(categorySection);
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