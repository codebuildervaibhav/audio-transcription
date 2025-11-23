# Diagnostic script - shows you the exact error

$env:GOROOT = "D:\go"
cd D:\Development\listner

Write-Host "Attempting to run server..." -ForegroundColor Yellow
Write-Host ""

go run cmd\server\main.go

Write-Host ""
Write-Host "If you see errors above, copy and paste them so I can fix them!" -ForegroundColor Cyan
