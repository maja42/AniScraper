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

echo Performing minimal build
echo ************************

if exist resources goto resourceDirExists
	echo Creating resources directory
	mkdir resources
:resourceDirExists

if exist export goto exportDirExists
	echo Creating export directory
	mkdir export
:exportDirExists


go build 											|| goto :error

move /y AniScraper.exe export\AniScraper.exe 		|| goto :error
xcopy /s/e/h/k/y resources export\resources\ 		|| goto :error

goto :EOF
:error
echo Failed with error #%errorlevel%.
exit /b %errorlevel%
