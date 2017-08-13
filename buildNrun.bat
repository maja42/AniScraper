@echo off

rem Bypass "Terminate Batch Job" prompt.
if "%~1"=="-FIXED_CTRL_C" (
   REM Remove the -FIXED_CTRL_C parameter
   SHIFT
) ELSE (
   REM Run the batch with <NUL and -FIXED_CTRL_C
   CALL <NUL %0 -FIXED_CTRL_C %*
   GOTO :EOF
)

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
exit /b %errorlevel%