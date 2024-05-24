#!/bin/bash

# Check if the migration file name is provided as an argument
if [ $# -eq 0 ]; then
  echo "Please provide the migration file name as an argument."
  exit 1
fi

# Get the current timestamp in the format YYYYMMDDHHMM
timestamp=$(date +%Y%m%d%H%M)

# Create the file name by combining the timestamp and the provided migration file name
filename="${timestamp}_$1.go"
directory="pkg/db/migrations/"

# Create the migration file in the specified directory
touch "$directory$filename"

echo "Migration file '$filename' created successfully."
