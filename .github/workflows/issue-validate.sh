# Fetch the latest changes from the main branch
set -ex
git fetch origin main

# Get the list of changed files
git diff origin/main...HEAD --name-only > diff

# Specify the directory containing migration files
MIGRATION_DIR="scripts/sql"

# Print changed files
echo "Changed files:"
cat diff

# Extract relevant .up.sql migration files from the changed files list
changed_files=$(grep "^$MIGRATION_DIR/" diff | grep -E "\.up\.sql$")

# Check if there are any .up.sql migration files in the changed files list
if [ -z "$changed_files" ]; then
  echo "No .up.sql migration files found in the changes."
  exit 0
fi
cd ..
cd ..
# Extract unique migration numbers from the directory (considering only .up.sql files)
existing_migrations=$(ls $MIGRATION_DIR | grep -E "\.up\.sql$" | grep -oE "[0-9]{3}[0-9]{3}[0-9]{2}" | sort | uniq)

# Exclude migration numbers from changed files in existing_migrations
while read -r file; do
  migration_number=$(basename "$file" | grep -oE "[0-9]{3}[0-9]{3}[0-9]{2}")
  existing_migrations=$(echo "$existing_migrations" | grep -v "$migration_number")
done <<< "$changed_files"

# Validate each changed .up.sql migration file
is_valid=true
processed_migrations=()
while read -r file; do
  # Extract migration number from the filename
  migration_number=$(basename "$file" | grep -oE "[0-9]{3}[0-9]{3}[0-9]{2}")

  if [ -z "$migration_number" ]; then
    echo "Warning: Could not extract migration number from $file."
    continue
  fi

  # Check if this migration number has already been processed
  if [[ " ${processed_migrations[@]} " =~ " $migration_number " ]]; then
    continue
  fi
  processed_migrations+=("$migration_number")

  # Check if the migration number is unique
  if echo "$existing_migrations" | grep -q "$migration_number"; then
    echo "Error: Migration number $migration_number already exists."
    is_valid=false
  fi

  # Check if the migration number is greater than previous ones
  last_migration=$(echo "$existing_migrations" | tail -n 1)
  if [ "$migration_number" -le "$last_migration" ]; then
    echo "Error: Migration number $migration_number is not greater than the latest ($last_migration)."
    is_valid=false
  fi

  # Check for sequential hotfix requirement (if NN > 01, check for NN-1)
  hotfix_number=$(echo "$migration_number" | grep -oE "[0-9]{2}$")
  if [ "$hotfix_number" -gt "01" ]; then
    previous_hotfix=$(printf "%02d" $((10#$hotfix_number - 1)))
    expected_previous_number="${migration_number:0:6}$previous_hotfix"
    if ! echo "$existing_migrations" | grep -q "$expected_previous_number"; then
      echo "Error: Previous hotfix migration $expected_previous_number not found for $migration_number."
      is_valid=false
    fi
  fi

done <<< "$changed_files"

if [ "$is_valid" = false ]; then
  echo "Validation failed. Please fix the errors before merging."
  exit 1
fi

echo "All .up.sql migration file validations passed."
