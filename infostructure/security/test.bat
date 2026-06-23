@echo off
chcp 65001 > nul

for /l %%i in (1,1,500) do (
    start /b curl -s -o NUL http://localhost:8080
)

echo end
pause
