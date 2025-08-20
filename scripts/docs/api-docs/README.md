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
