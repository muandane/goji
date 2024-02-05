#!/bin/sh

# Path to your Goji binary
GOJI_BINARY=$(go run . check)

# Get the commit message
COMMIT_MSG=$(git log --format=%B -n 1 $1)

# Run Goji with the commit message as argument
RESULT=$(go run . check "$COMMIT_MSG")

# Exit with non-zero status if Goji indicates a failed check
if [ $? -ne 0 ]; then
  echo "Commit message did not pass Goji check:"
  echo "$RESULT"
  exit 1
fi
# Otherwise, allow the commit to proceed
exit 0
