@echo off

rem Kill running application
tasklist /FI "IMAGENAME eq AniScraper.exe" 2>NUL | find /I /N "AniScraper.exe">NUL
if "%ERRORLEVEL%"=="0" taskkill /f /im AniScraper.exe

echo "Building..."
call build.bat || goto :error

echo Executing application
echo *************************************
call export\AniScraper.exe || goto :error

goto :EOF
:error
echo Failed with error #%errorlevel%.
pause
exit /b %errorlevel%