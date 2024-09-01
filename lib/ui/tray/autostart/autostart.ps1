param(
    [switch]$IsEnabled,
    [switch]$Enable,
    [switch]$Disable,
    [String]$TrayPath
)
$keyName = 'Timetracker Tray'
if ($IsEnabled) {
    $autostartProperty = Get-ItemProperty HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run -Name $keyName -ErrorAction SilentlyContinue
    if ($null -eq $autostartProperty) {
        exit 1
    }
    exit 0
}
if ($Enable) {
    if ($TrayPath -eq '') {
        throw 'TrayPath cannot be empty'
    }
    if (-not(Test-Path $TrayPath)) {
        throw "TrayPath '${TrayPath}' does not exist"
    }
    $autostartProperty = New-ItemProperty HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run -Name $keyName -Type String -Value $TrayPath -Force -ErrorAction SilentlyContinue
    if ($null -eq $autostartProperty) {
        exit 1
    }
    exit 0
}
if ($Disable) {
    Remove-ItemProperty HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Run -Name $keyName -Force
}
