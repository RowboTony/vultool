# PowerShell build script for vultool on Windows
# This script reads the version from the VERSION file and builds the binary

[CmdletBinding()]
param(
    [Parameter(HelpMessage="Build output path")]
    [string]$Output = "vultool.exe",
    
    [Parameter(HelpMessage="Additional build flags")]
    [string[]]$BuildFlags = @()
)

# Check if VERSION file exists
if (-not (Test-Path "VERSION")) {
    Write-Error "VERSION file not found in current directory"
    exit 1
}

# Read version from file
$version = Get-Content VERSION -Raw -ErrorAction Stop
$version = $version.Trim()

Write-Host "Building vultool version: $version" -ForegroundColor Green

# Build the binary
$ldflags = "-X main.version=$version"
$buildArgs = @(
    "build",
    "-ldflags", $ldflags
)

# Add any additional build flags
if ($BuildFlags) {
    $buildArgs += $BuildFlags
}

# Add output and source
$buildArgs += @("-o", $Output, "./cmd/vultool")

Write-Host "Running: go $($buildArgs -join ' ')" -ForegroundColor Cyan

# Execute the build
& go $buildArgs

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful! Binary created: $Output" -ForegroundColor Green
    
    # Test the binary
    Write-Host "`nTesting binary..." -ForegroundColor Yellow
    & "./$Output" --version
} else {
    Write-Error "Build failed with exit code: $LASTEXITCODE"
    exit $LASTEXITCODE
}
