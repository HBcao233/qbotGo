chcp 65001
Start-Process -FilePath "go.exe" -ArgumentList "run","main.go" -WindowStyle Hidden -PassThru -RedirectStandardOutput bot.log | Format-List Name,Id,Path
