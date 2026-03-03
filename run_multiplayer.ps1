param (
    [int]$count = 2
)

$godotPath = "C:\Users\kevin\Downloads\godot\godot.exe"
$projectDir = "c:\Users\kevin\SynologyDrive\Godot\Kashino\frontend"

Write-Host "Launching $count multiplayer instances..." -ForegroundColor Yellow

for ($i = 1; $i -le $count; $i++) {
    Write-Host "Launching Player $i..." -ForegroundColor Green
    Start-Process $godotPath -ArgumentList "--path `"$projectDir`" -- --profile=$i --side-by-side=$count"
}
