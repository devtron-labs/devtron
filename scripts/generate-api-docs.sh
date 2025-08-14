#!/bin/bash

# Script to generate HTML documentation from all API specs using Redocly
# This will convert all .yaml and .yml files in the specs directory to HTML

set -ex

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

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if redocly is installed
if ! command -v redocly &> /dev/null; then
    echo -e "${RED}❌ Redocly is not installed. Please install it first:${NC}"
    echo "npm install -g @redocly/cli"
    exit 1
fi

echo -e "${GREEN}✅ Redocly found: $(redocly --version)${NC}"

# Counter for successful conversions
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

    # Create output filename
    local output_file="$OUTPUT_DIR/${relative_path%.*}.html"

    # Create output directory if it doesn't exist
    local output_dir
    output_dir=$(dirname "$output_file")
    mkdir -p "$output_dir"

    echo -e "${BLUE}📄 Converting: $spec_file${NC}"

    # Capture both stdout and stderr to check for errors
    if redocly build-docs "$spec_file" -o "$output_file" 2>&1; then
        echo -e "${GREEN}✅ Success: $output_file${NC}"
        ((success_count++))
        return 0
    else
        echo -e "${RED}❌ Failed: $spec_file${NC}"
        echo "Failed: $spec_file" >> "$ERROR_LOG"
        ((error_count++))
        return 1
    fi
}

# Find all spec files and convert them
echo -e "${YELLOW}🔍 Finding all spec files...${NC}"
spec_files=$(find "$SPECS_DIR" -name "*.yaml" -o -name "*.yml" | sort)

echo -e "${BLUE}📊 Found $(echo "$spec_files" | wc -l | tr -d ' ') spec files${NC}"

# Convert each spec file
for spec_file in $spec_files; do
    convert_spec_to_html "$spec_file" || true
done

# Generate index page
echo -e "${YELLOW}📝 Generating index page...${NC}"

cat > "$INDEX_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Devtron API Documentation</title>
    <style>
        /* styles omitted for brevity — keep your existing CSS here */
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Devtron API Documentation</h1>
        <div class="description">
            Comprehensive API documentation for Devtron - Kubernetes-native software delivery platform
        </div>

        <div class="categories" id="categories">
            <!-- Categories will be populated by JavaScript -->
        </div>

        <div class="footer">
            <p><a href="/https://devtron.ai/" target="_blank">Devtron</a></p>
            <p class="timestamp">Last updated: <span id="timestamp"></span></p>
        </div>
    </div>

    <script>
        const apiData = {
EOF

# Generate the JavaScript data for the index page
echo "Generating JavaScript data for index page..."

total_apis=0
categories_count=0

for spec_file in $spec_files; do
    relative_path="${spec_file#$SPECS_DIR/}"
    filename=$(basename "$spec_file")
    name_without_ext="${filename%.*}"
    category=$(dirname "$relative_path")

    if [ "$category" = "." ]; then
        category="root"
    fi

    display_category=$(echo "$category" | sed 's/-/ /g' | sed 's/_/ /g' | sed 's/\b\w/\U&/g')

    title=$(grep -m 1 '^[[:space:]]*title:' "$spec_file" \
        | sed 's/^[[:space:]]*title:[[:space:]]*//' \
        | tr -d '"' \
        || echo "$name_without_ext")

    output_file="${relative_path%.*}.html"

    if [ -f "$OUTPUT_DIR/$output_file" ]; then
        cat >> "$INDEX_FILE" << EOF
            "${category}_${name_without_ext}": {
                "category": "${display_category}",
                "title": "${title}",
                "filename": "${output_file}",
                "original_file": "${relative_path}"
            },
EOF
    fi
done

# Log what find returns for index.html
echo -e "${YELLOW}🔍 Searching for index.html in $OUTPUT_DIR...${NC}"
find . -name "index.html" -path "*/docs/api-docs/*"

# Remove trailing comma only if file exists
if [[ -f "$INDEX_FILE" ]]; then
    sed -i '$ s/,$//' "$INDEX_FILE"
else
    echo -e "${RED}⚠️ $INDEX_FILE not found, skipping trailing comma fix${NC}"
fi

cat >> "$INDEX_FILE" << 'EOF'
        };

        function populatePage() {
            const categoriesContainer = document.getElementById('categories');
            const categories = {};

            Object.values(apiData).forEach(api => {
                if (!categories[api.category]) {
                    categories[api.category] = [];
                }
                categories[api.category].push(api);
            });

            Object.keys(categories).sort().forEach(category => {
                const categoryDiv = document.createElement('div');
                categoryDiv.className = 'category';

                const categoryTitle = document.createElement('h3');
                categoryTitle.textContent = category;
                categoryDiv.appendChild(categoryTitle);

                const apiList = document.createElement('ul');
                apiList.className = 'api-list';

                categories[category].sort((a, b) => a.title.localeCompare(b.title)).forEach(api => {
                    const listItem = document.createElement('li');
                    const link = document.createElement('a');
                    link.href = api.filename;
                    link.textContent = api.title;
                    link.target = '_blank';
                    listItem.appendChild(link);
                    apiList.appendChild(listItem);
                });

                categoryDiv.appendChild(apiList);
                categoriesContainer.appendChild(categoryDiv);
            });

            document.getElementById('timestamp').textContent = new Date().toLocaleString();
        }

        document.addEventListener('DOMContentLoaded', populatePage);
    </script>
</body>
</html>
EOF

echo -e "${GREEN}✅ Index page generated: $INDEX_FILE${NC}"

# Final summary
echo -e "${BLUE}📊 Final Summary:${NC}"
echo -e "${GREEN}✅ Successfully converted: $success_count files${NC}"
if [ $error_count -gt 0 ]; then
    echo -e "${RED}❌ Failed to convert: $error_count files${NC}"
    echo -e "${YELLOW}📝 Check $ERROR_LOG for details on failed conversions${NC}"
fi
echo -e "${BLUE}📁 Output directory: $OUTPUT_DIR${NC}"
echo -e "${BLUE}🌐 Main index: $INDEX_FILE${NC}"

# Create README
cat > "$OUTPUT_DIR/README.md" << 'EOF'
# Devtron API Documentation
... (keep your existing README text here)
EOF

echo -e "${GREEN}✅ README created: $OUTPUT_DIR/README.md${NC}"
echo -e "${GREEN}🎉 API documentation generation complete!${NC}"
echo -e "${BLUE}💡 Open $INDEX_FILE in your browser to view the documentation${NC}"
