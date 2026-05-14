#Requires -Version 5.1
# Windows 전용: Air로 로컬 개발 서버 실행 (mac/Linux는 scripts/dev.sh + npm run dev)
$ErrorActionPreference = 'Stop'

$ROOT = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
Set-Location -LiteralPath $ROOT

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    foreach ($d in @("${env:ProgramFiles}\Go\bin", "${env:ProgramFiles(x86)}\Go\bin")) {
        if (Test-Path -LiteralPath (Join-Path $d 'go.exe')) {
            $env:PATH = "$d;$env:PATH"
            break
        }
    }
}

$port = $env:PORT
if (-not $port -and (Test-Path -LiteralPath (Join-Path $ROOT '.env'))) {
    $line = Get-Content -LiteralPath (Join-Path $ROOT '.env') -ErrorAction SilentlyContinue |
        Where-Object { $_ -match '^\s*PORT\s*=' } |
        Select-Object -First 1
    if ($line) {
        $port = ($line -split '=', 2)[1].Trim().Trim('"').Trim("'")
    }
}
if (-not $port) { $port = '8080' }

$portNum = 8080
if (-not [int]::TryParse("$port".Trim(), [ref]$portNum)) {
    $portNum = 8080
}

$pids = @(Get-NetTCPConnection -LocalPort $portNum -State Listen -ErrorAction SilentlyContinue |
        ForEach-Object { $_.OwningProcess } |
        Sort-Object -Unique)
foreach ($p in $pids) {
    if ($p -and $p -gt 0) {
        Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
    }
}
Start-Sleep -Milliseconds 300
$pids2 = @(Get-NetTCPConnection -LocalPort $portNum -State Listen -ErrorAction SilentlyContinue |
        ForEach-Object { $_.OwningProcess } |
        Sort-Object -Unique)
foreach ($p in $pids2) {
    if ($p -and $p -gt 0) {
        Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
    }
}

if ($env:AIR -and (Test-Path -LiteralPath $env:AIR)) {
    & $env:AIR
    exit $LASTEXITCODE
}

$airGlobal = Get-Command air -ErrorAction SilentlyContinue
if ($airGlobal) {
    & $airGlobal.Source
    exit $LASTEXITCODE
}

$BIN_DIR = Join-Path $ROOT 'bin'
New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null
$AIRBIN = Join-Path $BIN_DIR 'air.exe'
if (-not (Test-Path -LiteralPath $AIRBIN)) {
    Write-Host "dev: Air를 $AIRBIN 에 설치합니다(최초 1회)." -ForegroundColor Yellow
    $env:GOBIN = $BIN_DIR
    go install github.com/air-verse/air@latest
}

& $AIRBIN
exit $LASTEXITCODE
