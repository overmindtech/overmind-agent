$ComputerHW = Get-WmiObject -Class Win32_ComputerSystem |
    Select-Object Manufacturer,Model | Format-Table -AutoSize

$ComputerCPU = Get-WmiObject win32_processor |
    Select-Object DeviceID,Name | Format-Table -AutoSize

$ComputerRam_Total = Get-WmiObject Win32_PhysicalMemoryArray |
    Select-Object MemoryDevices,MaxCapacity | Format-Table -AutoSize

$ComputerRAM = Get-WmiObject Win32_PhysicalMemory |
    Select-Object DeviceLocator,Manufacturer,PartNumber,Capacity,Speed | Format-Table -AutoSize

$ComputerDisks = Get-WmiObject -Class Win32_LogicalDisk -Filter "DriveType=3" |
    Select-Object DeviceID,VolumeName,Size,FreeSpace | Format-Table -AutoSize

$ComputerOS = (Get-WmiObject Win32_OperatingSystem).Version

switch -Wildcard ($ComputerOS) {
    "6.1.7600" {$OS = "Windows 7"; break}
    "6.1.7601" {$OS = "Windows 7 SP1"; break}
    "6.2.9200" {$OS = "Windows 8"; break}
    "6.3.9600" {$OS = "Windows 8.1"; break}
    "10.0.*" {$OS = "Windows 10"; break}
    default {$OS = "Unknown Operating System"; break}
}

Write-Host "Computer Name: $Computer"
Write-Host "Operating System: $OS"
Write-Output $ComputerHW
Write-Output $ComputerCPU
Write-Output $ComputerRam_Total
Write-Output $ComputerRAM
Write-Output $ComputerDisks
