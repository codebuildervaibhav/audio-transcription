# Fix Go Installation - Run in PowerShell as Administrator

Write-Host "Fixing Go Environment Variables..." -ForegroundColor Yellow
Write-Host ""

# The issue: GOROOT is pointing to C:\Program Files\Go\bin
# Should be: D:\go

# Fix GOROOT
$goRoot = "D:\go"

if (Test-Path $goRoot) {
    Write-Host "[OK] Go installation found at: $goRoot" -ForegroundColor Green
    
    # Set system environment variable
    [System.Environment]::SetEnvironmentVariable("GOROOT", $goRoot, "Machine")
    
    # Update current session
    $env:GOROOT = $goRoot
    
    # Ensure Go bin is in PATH
    $goBin = Join-Path $goRoot "bin"
    $currentPath = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
    
    if ($currentPath -notlike "*$goBin*") {
        Write-Host "Adding Go to PATH..." -ForegroundColor Yellow
        [System.Environment]::SetEnvironmentVariable(
            "Path",
            "$currentPath;$goBin",
            "Machine"
        )
    }
    
    # Update current session PATH
    $env:Path = "$env:Path;$goBin"
    
    Write-Host "[OK] GOROOT set to: $goRoot" -ForegroundColor Green
    Write-Host "[OK] Go bin added to PATH: $goBin" -ForegroundColor Green
    Write-Host ""
    Write-Host "Testing Go installation..." -ForegroundColor Yellow
    
    # Refresh environment in current session
    $env:GOROOT = $goRoot
    
    # Test
    & "$goBin\go.exe" version
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "[SUCCESS] Go is now working!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Please CLOSE and REOPEN PowerShell for changes to take full effect" -ForegroundColor Cyan
        Write-Host "Then run: go version" -ForegroundColor White
    }
    
} else {
    Write-Host "[ERROR] Go not found at $goRoot" -ForegroundColor Red
    Write-Host "Please install Go from: https://golang.org/dl/" -ForegroundColor Yellow
}

Write-Host ""
