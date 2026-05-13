param([string]$T="complete")
if($T -eq "complete"){[Console]::Beep(800,150);Start-Sleep -m 100;[Console]::Beep(800,150);Start-Sleep -m 100;[Console]::Beep(1000,400)}
elseif($T -eq "error"){[Console]::Beep(200,500);Start-Sleep -m 200;[Console]::Beep(150,500)}
