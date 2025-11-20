# PowerShell script to merge all test case files into all.txt

# Set the working directory to the script location
Set-Location -Path $PSScriptRoot

# Define the output file
$outputFile = "all.txt"

# Clear the output file if it exists
if (Test-Path $outputFile) {
    Clear-Content -Path $outputFile
}

# Get all txt files except all.txt
$files = Get-ChildItem -Path "*.txt" -Exclude "all.txt"

# Merge the contents into all.txt
foreach ($file in $files) {
    # Add the file contents, filtering out empty lines and lines with only whitespace
    Get-Content -Path $file.FullName | Where-Object {
        $_.Trim() -ne ""
    } | Add-Content -Path $outputFile
}

Write-Host "Merged all test case files into $outputFile"
