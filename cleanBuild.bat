@echo off

echo Performing clean build
echo **********************

if exist resources goto resourceDirExists
	echo Creating resources directory
	mkdir resources
:resourceDirExists

RMDIR /S /Q export
MKDIR export

go build 							            || goto :error

move /y AniScraper.exe export\AniScraper.exe    || goto :error
xcopy /s/e/h/k/y resources export\resources\    || goto :error


goto :EOF
:error
echo Failed with error #%errorlevel%.
pause
exit /b %errorlevel%
