#!/bin/sh

# The first argument is the path to the temporary file that contains the commit message.
COMMIT_MSG_FILE=$1

# Read the commit message into a variable.
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Run your 'goji check' command against the commit message.
if ! echo "$COMMIT_MSG" | goji check; then
  # If 'goji check' returns a non-zero exit code, abort the commit.
  echo "Aborting commit due to failed 'goji check'"
  exit 1
fi

# If everything is OK, allow the commit to proceed.
exit 0
