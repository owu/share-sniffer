#!/bin/bash

# Bash script to merge all test case files into all.txt

# Set the working directory to the script location
cd "$(dirname "$0")"

# Define the output file
OUTPUT_FILE="all.txt"

# Clear the output file if it exists
> "$OUTPUT_FILE"

# Get all txt files except all.txt
for file in *.txt; do
    if [ "$file" != "$OUTPUT_FILE" ]; then
        # Add the file contents, filtering out empty lines and lines with only whitespace
        grep -v '^[[:space:]]*$' "$file" >> "$OUTPUT_FILE"
    fi
done

echo "Merged all test case files into $OUTPUT_FILE"
