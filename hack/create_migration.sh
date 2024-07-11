#!/bin/bash


if [ $# -eq 0 ]; then
  echo "Please provide the migration file name as an argument."
  exit 1
fi


timestamp=$(date +%Y%m%d%H%M)


filename="${timestamp}_$1.go"
directory="pkg/db/migrations/"

# Create the migration file in the specified directory
touch "$directory$filename"

echo "Migration file '$filename' created successfully."
