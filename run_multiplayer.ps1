$godotPath = "C:\Users\kevin\Downloads\godot\godot.exe"
$projectDir = "c:\Users\kevin\SynologyDrive\Godot\Kashino\frontend"

Write-Host "Launching Player 1..." -ForegroundColor Green
Start-Process $godotPath -ArgumentList "--path `"$projectDir`" -- --profile=1"

Write-Host "Launching Player 2..." -ForegroundColor Cyan
Start-Process $godotPath -ArgumentList "--path `"$projectDir`" -- --profile=2"
