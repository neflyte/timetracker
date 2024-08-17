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
    [Parameter(Mandatory)][string]$AppId,
    [Parameter(Mandatory)][string]$Title,
    [string]$Description,
    [string]$Icon,
    [string]$Launch
)
if ($PSVersionTable.PSVersion.Major -gt 5) {
    throw 'This script is only supported in PowerShell 5.1'
}
# Debug the script parameters
Write-Debug "AppId: $AppId, Title: $Title, Description: $Description, Icon: $Icon, Launch: $Launch"
$launchProtocol = $Launch
if ($launchProtocol -eq '') {
    $launchProtocol = 'timetracker:main'
}
# Create the notification configuration XML
$doc = [xml]::new()
$bindingEl = $doc.CreateElement('binding')
$null = $bindingEl.SetAttribute('template','ToastGeneric')
if ($Icon -ne '') {
    $iconEl = $doc.CreateElement('image')
    $null = $iconEl.SetAttribute('placement','appLogoOverride')
    $null = $iconEl.SetAttribute('src',$Icon)
    $null = $bindingEl.AppendChild($iconEl)
}
$titleEl = $doc.CreateElement('text')
$titleCData = $doc.CreateCDataSection($Title)
$null = $titleEl.AppendChild($titleCData)
$null = $bindingEl.AppendChild($titleEl)
if ($Description -ne '') {
    $descEl = $doc.CreateElement('text')
    $descCData = $doc.CreateCDataSection($Description)
    $null = $descEl.AppendChild($descCData)
    $null = $bindingEl.AppendChild($descEl)
}
$visualEl = $doc.CreateElement('visual')
$null = $visualEl.AppendChild($bindingEl)
$audioEl = $doc.CreateElement('audio')
$null = $audioEl.SetAttribute('src','ms-winsoundevent:Notification.Default')
$null = $audioEl.SetAttribute('loop','false')
$actionsEl = $doc.CreateElement('actions')
$toastEl = $doc.CreateElement('toast')
$null = $toastEl.SetAttribute('activationType','protocol')
$null = $toastEl.SetAttribute('launch',$launchProtocol)
$null = $toastEl.SetAttribute('duration','short')
$null = $toastEl.AppendChild($visualEl)
$null = $toastEl.AppendChild($audioEl)
$null = $toastEl.AppendChild($actionsEl)
$null = $doc.AppendChild($toastEl)
# Debug the notification cofiguration XML
Write-Debug "doc:`n$($doc.OuterXml)"
# Send the notification
$null = [Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime]
$null = [Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime]
$null = [Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime]
$toastXml = [Windows.Data.Xml.Dom.XmlDocument]::new()
$toastXml.LoadXml($doc.OuterXml)
$toast = [Windows.UI.Notifications.ToastNotification]::new($toastXml)
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($AppId).Show($toast)
