# Start Server - Simplified version

$env:GOROOT = "D:\go"
$env:Path += ";C:\Users\vaibh\AppData\Local\Programs\Python\Python313\Scripts"

cd D:\Development\listner

Write-Host "Starting server..." -ForegroundColor Cyan
go run cmd\server\main.go
