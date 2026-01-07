# @description Install pingmonke as a Windows Service
$serviceName = "Pingmonke"
$exePath = "C:\Program Files\Pingmonke\pingmonke.exe"

New-Service -Name $serviceName -BinaryPathName $exePath -StartupType Automatic
Start-Service -Name $serviceName