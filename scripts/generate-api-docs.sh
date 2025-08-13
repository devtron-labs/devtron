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
    local output_dir=$(dirname "$output_file")
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
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            text-align: center;
            margin-bottom: 30px;
            font-size: 2.5em;
        }
        .description {
            text-align: center;
            color: #7f8c8d;
            margin-bottom: 40px;
            font-size: 1.1em;
        }
        .stats {
            background: #ecf0f1;
            padding: 20px;
            border-radius: 6px;
            margin-bottom: 30px;
            text-align: center;
        }
        .stats h2 {
            margin: 0 0 15px 0;
            color: #34495e;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-top: 15px;
        }
        .stat-item {
            background: white;
            padding: 15px;
            border-radius: 6px;
            border-left: 4px solid #3498db;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #3498db;
        }
        .stat-label {
            color: #7f8c8d;
            font-size: 0.9em;
        }
        .categories {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
        }
        .category {
            background: #f8f9fa;
            border-radius: 6px;
            padding: 20px;
            border: 1px solid #e9ecef;
        }
        .category h3 {
            color: #2c3e50;
            margin-top: 0;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        .api-list {
            list-style: none;
            padding: 0;
        }
        .api-list li {
            margin: 8px 0;
        }
        .api-list a {
            color: #3498db;
            text-decoration: none;
            padding: 5px 10px;
            border-radius: 4px;
            transition: background-color 0.2s;
            display: inline-block;
        }
        .api-list a:hover {
            background-color: #e3f2fd;
            text-decoration: underline;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ecf0f1;
            color: #7f8c8d;
        }
        .timestamp {
            font-size: 0.9em;
            color: #95a5a6;
        }
        .errors-section {
            background: #fff5f5;
            border: 1px solid #fed7d7;
            border-radius: 6px;
            padding: 20px;
            margin: 20px 0;
        }
        .errors-section h3 {
            color: #c53030;
            margin-top: 0;
        }
        .error-list {
            background: white;
            border-radius: 4px;
            padding: 15px;
            font-family: monospace;
            font-size: 0.9em;
        }
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
        // API data structure
        const apiData = {
EOF

# Generate the JavaScript data for the index page
echo "Generating JavaScript data for index page..."

# Initialize counters
total_apis=0
categories_count=0

# Process each spec file to build the data structure
for spec_file in $spec_files; do
    relative_path="${spec_file#$SPECS_DIR/}"
    filename=$(basename "$spec_file")
    name_without_ext="${filename%.*}"
    category=$(dirname "$relative_path")

    # Skip if it's the root specs directory
    if [ "$category" = "." ]; then
        category="root"
    fi

    # Clean up category name for display
    display_category=$(echo "$category" | sed 's/-/ /g' | sed 's/_/ /g' | sed 's/\b\w/\U&/g')

    # Get the title from the spec file (first line with 'title:')
    title=$(grep -m 1 '^[[:space:]]*title:' "$spec_file" | sed 's/^[[:space:]]*title:[[:space:]]*//' | tr -d '"' || echo "$name_without_ext")

    # Create the output filename
    output_file="${relative_path%.*}.html"

    # Check if the HTML file was actually created successfully
    if [ -f "$OUTPUT_DIR/$output_file" ]; then
        # Add to JavaScript data
        cat >> "$INDEX_FILE" << EOF
            "${category}_${name_without_ext}": {
                "category": "${display_category}",
                "title": "${title}",
                "filename": "${output_file}",
                "original_file": "${relative_path}"
            },
EOF
        ((total_apis++))
    fi
done

# Remove the last comma and close the data structure
sed -i '' '$ s/,$//' "$INDEX_FILE"

cat >> "$INDEX_FILE" << 'EOF'
        };

        // Function to populate the page
        function populatePage() {
            const categoriesContainer = document.getElementById('categories');
            const categories = {};

            // Group APIs by category
            Object.values(apiData).forEach(api => {
                if (!categories[api.category]) {
                    categories[api.category] = [];
                }
                categories[api.category].push(api);
            });

            // Create category sections
            Object.keys(categories).sort().forEach(category => {
                const categoryDiv = document.createElement('div');
                categoryDiv.className = 'category';

                const categoryTitle = document.createElement('h3');
                categoryTitle.textContent = category;
                categoryDiv.appendChild(categoryTitle);

                const apiList = document.createElement('ul');
                apiList.className = 'api-list';

                // Sort APIs within category by title
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

            // Update statistics

            document.getElementById('timestamp').textContent = new Date().toLocaleString();
        }

        // Initialize the page
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

# Create a simple README for the docs
cat > "$OUTPUT_DIR/README.md" << 'EOF'
# Devtron API Documentation

This directory contains HTML documentation generated from OpenAPI specifications using Redocly.

## Files

- `index.html` - Main index page with links to all API documentation
- Individual HTML files for each API specification
- `errors.log` - Log of any files that failed to convert

## How to Use

1. Open `index.html` in your web browser to see the main index
2. Click on any API link to view the detailed documentation
3. All documentation is self-contained and can be hosted on any web server

## Generation

To regenerate the documentation, run:

```bash
./scripts/generate-api-docs.sh
```

## Requirements

- Redocly CLI (`npm install -g @redocly/cli`)
- Bash shell

## Notes

- Each HTML file is self-contained and includes all necessary CSS and JavaScript
- Documentation is generated from OpenAPI 3.0+ specifications
- Files are organized by category for easy navigation
- If some files fail to convert, check the errors.log file for details
EOF

echo -e "${GREEN}✅ README created: $OUTPUT_DIR/README.md${NC}"
echo -e "${GREEN}🎉 API documentation generation complete!${NC}"
echo -e "${BLUE}💡 Open $INDEX_FILE in your browser to view the documentation${NC}"
