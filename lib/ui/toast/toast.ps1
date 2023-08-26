<#
.SYNOPSIS
    toast.ps1 sends a toast notification
.PARAMETER AppId
    The application ID
.PARAMETER Title
    The first line of the notification
.PARAMETER Description
    The second line of the notification
.PARAMETER Icon
    Path and filename of the notification icon
.PARAMETER Launch
    The protocol launch string
#>
param(
    [Parameter(Mandatory=$true)][string]$AppId,
    [Parameter(Mandatory=$true)][string]$Title,
    [string]$Description = '',
    [string]$Icon = '',
    [string]$Launch = 'timetracker:main'
)
if ($PSVersionTable.PSVersion.Major -gt 5) {
    Write-Error 'This script is only supported in PowerShell 5.1'
    throw 'unsupported PowerShell version'
}
# Debug the script parameters
Write-Debug "AppId: $AppId, Title: $Title, Description: $Description, Icon: $Icon, Launch: $Launch"
# Create the notification configuration XML
$doc = [xml]::new()
$bindingEl = $doc.CreateElement('binding')
$bindingEl.SetAttribute('template','ToastGeneric')
if ($Icon -ne '') {
    $iconEl = $doc.CreateElement('image')
    $iconEl.SetAttribute('placement','appLogoOverride')
    $iconEl.SetAttribute('src',$Icon)
    $bindingEl.AppendChild($iconEl)
}
$titleEl = $doc.CreateElement('text')
$titleEl.AppendChild($doc.CreateCDataSection($Title))
$bindingEl.AppendChild($titleEl)
if ($Description -ne '') {
    $descEl = $doc.CreateElement('text')
    $descEl.AppendChild($doc.CreateCDataSection($Description))
    $bindingEl.AppendChild($descEl)
}
$visualEl = $doc.CreateElement('visual')
$visualEl.AppendChild($bindingEl)
$audioEl = $doc.CreateElement('audio')
$audioEl.SetAttribute('src','ms-winsoundevent:Notification.Default')
$audioEl.SetAttribute('loop','false')
$actionsEl = $doc.CreateElement('actions')
$toastEl = $doc.CreateElement('toast')
$toastEl.SetAttribute('activationType','protocol')
$toastEl.SetAttribute('launch',$Launch)
$toastEl.SetAttribute('duration','short')
$toastEl.AppendChild($visualEl)
$toastEl.AppendChild($audioEl)
$toastEl.AppendChild($actionsEl)
$doc.AppendChild($toastEl)
# Debug the notification cofiguration XML
Write-Debug "doc:`n$($doc.InnerXml)"
# Send the notification
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
$toastXml = New-Object Windows.Data.Xml.Dom.XmlDocument
$toastXml.LoadXml($doc.InnerXml)
$toast = New-Object Windows.UI.Notifications.ToastNotification $toastXml
$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($AppId)
$notifier.Show($toast)
