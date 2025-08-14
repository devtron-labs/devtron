#!/bin/bash
set -e

OUTPUT_DIR="docs/api-docs"
INDEX_FILE="$OUTPUT_DIR/index.html"

# 1. Clean output dir
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# 2. Prepare index.html
cat > "$INDEX_FILE" << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { border-bottom: 2px solid #ccc; padding-bottom: 10px; }
        .category { margin-bottom: 20px; }
        .category h3 { margin-top: 0; color: #2c3e50; }
        .api-list { list-style: none; padding-left: 0; }
        .api-list li { margin: 5px 0; }
        .timestamp { margin-top: 40px; font-size: 0.9em; color: #888; }
    </style>
</head>
<body>
    <h1>API Documentation</h1>
    <div id="categories"></div>
    <div class="timestamp">Last updated: <span id="timestamp"></span></div>
    <script>
        const apiData = {
EOF

# 3. Loop over all specs
find specs -type f -name "*.yaml" | while read -r spec; do
    title=$(grep -m1 "^title:" "$spec" | sed 's/^[[:space:]]*title:[[:space:]]*//' | tr -d '"')
    filename=$(basename "$spec" .yaml).html
    category=$(basename "$(dirname "$spec")")

    # Run redocly to generate HTML
    npx @redocly/cli build-docs "$spec" -o "$OUTPUT_DIR/$filename"

    # Add entry to index.html
    cat >> "$INDEX_FILE" << EOF
            "${filename}": {
                "title": "${title}",
                "category": "${category}",
                "filename": "${filename}"
            },
EOF
done

# 4. Remove last comma in apiData
sed -i '$ s/,$//' "$INDEX_FILE"

# 5. Append JS for rendering categories & APIs
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
                categoryTitle.textContent = category
                    .replace(/([a-z])([A-Z])/g, '$1 $2')  // camelCase -> space
                    .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2') // ABCThing -> ABC Thing
                    .split(/[\s-_]+/) // split on space, dash, underscore
                    .map(word => {
                        if (word === word.toUpperCase()) return word; // keep acronyms uppercase
                        return word.charAt(0).toUpperCase() + word.slice(1);
                    })
                    .join(' ');

                categoryDiv.appendChild(categoryTitle);

                const apiList = document.createElement('ul');
                apiList.className = 'api-list';

                categories[category]
                    .sort((a, b) => a.title.localeCompare(b.title))
                    .forEach(api => {
                        const listItem = document.createElement('li');
                        const link = document.createElement('a');
                        link.href = api.filename; // removed target="_blank"
                        link.textContent = api.title;
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

echo "✅ API documentation generated at $OUTPUT_DIR"
