@echo off

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
