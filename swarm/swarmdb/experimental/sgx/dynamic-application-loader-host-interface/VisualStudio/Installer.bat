@echo off
REM - This script installs JHI in the current machin.
REM - Author : Elad Dabool
REM - For any issue, please email me at elad.dabool@intel.com

set is_X64_OS=N
if EXIST "C:\Program Files (x86)" set is_X64_OS=Y

if "%~2"=="emulation" ( 
	set emulation_extension=_EMULATION
	set sockets=true
) else (
  if "%~2"=="mei" (
  	set emulation_extension=
	set sockets=
  ) else (
  if "%~2"=="sockets" (
	set sockets=true
	set emulation_extension=
  ) else (
	call:PrintUsage
	EXIT /B 1
	)
  )
)

set DalPath_X86_OS=C:\Program Files\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%
set DalPath_X64_OS_PRIMERY=C:\Program Files (x86)\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%
set DalPath_X64_OS_SECONDERY=C:\Program Files\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%
set AppletsFolder=C:\ProgramData\Intel\DAL%emulation_extension%\Applets
set Command=%1%

call:PrintLogo

IF %Command%.==install. (
	goto MAIN.INSTALL
) else IF %Command%.==uninstall. (
	goto MAIN.UNINSTALL
) else IF %Command%.==start. (
	goto MAIN.START
) else IF %Command%.==stop. (
	goto MAIN.STOP
) else (
	call:PrintUsage
	EXIT /B 1
)

:EXIT_SUCCESS
EXIT /B 0

:EXIT_FAILURE
EXIT /B 1

REM - *************************     FUNCTIONS **************************

:MAIN.INSTALL
echo installing JHI...
echo.

call:VerifyAdminPrivileges
IF %ERRORLEVEL%==1 EXIT /B 1

call:VerifyJhiFilesExists
IF %ERRORLEVEL%==1 (
	echo.
	echo Error: some of JHI files are missing. aborting.
	goto EXIT_FAILURE
)

call:StopJHIService > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed to stop JHI Service. aborting.
	goto EXIT_FAILURE
)

call:UninstallJHIService > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed to unregister JHI Service. aborting.
	goto EXIT_FAILURE
)

call:RemoveJHIDirectories > NUL
IF %ERRORLEVEL%==1 (
	echo Error: Error: failed to remove JHI directories. aborting.
	goto EXIT_FAILURE
)

call:CreateJHIDirectories > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed to create JHI directories. aborting.
	goto EXIT_FAILURE
)

call:CopyJHIFiles > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed copying JHI Service files. aborting.
	goto EXIT_FAILURE
)

call:AddRegistryKeys > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed adding JHI keys in registry. aborting.
	goto EXIT_FAILURE
)

call:InstallJHIService > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed register JHI Service. aborting.
	goto EXIT_FAILURE
)

call:StartJHIService > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed starting JHI Service. aborting.
	goto EXIT_FAILURE
)

call:AddFoldersToPath.program
if not %errorlevel%==0 (
	echo Warning: failed adding JHI paths to PATH.
	echo It is recommended to update the path manually.
)

echo Install completed successfuly.
echo.
goto EXIT_SUCCESS

:MAIN.UNINSTALL
echo uninstalling JHI...
echo.

call:VerifyAdminPrivileges
IF %ERRORLEVEL%==1 EXIT /B 1

call:StopJHIService > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed to stop JHI Service. aborting.
	goto EXIT_FAILURE
)

call:UninstallJHIService > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed to unregister JHI Service. aborting.
	goto EXIT_FAILURE
)

call:RemoveRegistryKeys > NUL
IF %ERRORLEVEL%==1 (
	echo Error: failed remove JHI keys from registry. aborting.
	goto EXIT_FAILURE
)
call:RemoveJHIDirectories > NUL
IF %ERRORLEVEL%==1 (
	echo Error: Error: failed to remove JHI directories. aborting.
	goto EXIT_FAILURE
)

echo Uninstall completed successfuly.
echo.
echo **** DAL%emulation_extension% installation paths can be removed from the PATH Environment Variable ****
echo.

goto EXIT_SUCCESS

:MAIN.START
echo starting JHI...
echo.

call:VerifyAdminPrivileges
IF %ERRORLEVEL%==1 EXIT /B 1

call:StartJHIService 
IF %ERRORLEVEL% EQU 0 goto EXIT_SUCCESS
goto EXIT_FAILURE

:MAIN.STOP
echo stopping JHI...
echo.

call:VerifyAdminPrivileges
IF %ERRORLEVEL%==1 EXIT /B 1

call:StopJHIService 
IF %ERRORLEVEL% EQU 0 goto EXIT_SUCCESS
goto EXIT_FAILURE

:AddRegistryKeys
set JhiPath="C:\\Program Files (x86)\\Intel\\Intel(R) Management Engine Components\\DAL%emulation_extension%"

IF %is_X64_OS% EQU N (
	set JhiPath="C:\\Program Files\\Intel\\Intel(R) Management Engine Components\\DAL%emulation_extension%"
)

set AppletsPath=C:\\ProgramData\\Intel\\DAL%emulation_extension%\\Applets

set TempRegistryFile=add.reg

set TransportType=2 rem heci default
if [%sockets%]==[true] (
	rem sockets
	set TransportType=1
)

IF EXIST %TempRegistryFile% DEL %TempRegistryFile%

call:CreateAddRegistryFile %TempRegistryFile% %JhiPath% %AppletsPath% %TransportType%

REGEDIT /S %TempRegistryFile%

DEL %TempRegistryFile%

IF %ERRORLEVEL% EQU 0 EXIT /B 0
EXIT /B 1

:CreateAddRegistryFile
set TempFile=%1%
set JHIDir=%~2%
set AppletsDir=%~3%
set Transport_Type=%~4%

echo Windows Registry Editor Version 5.00 >> %TempFile%
echo. >> %TempFile%

echo [HKEY_LOCAL_MACHINE\SOFTWARE\Intel\Services\DAL%emulation_extension%] >> %TempFile%
echo "FILELOCALE"="%JHIDir%" >> %TempFile%
echo. >> %TempFile%

echo [HKEY_LOCAL_MACHINE\SOFTWARE\Intel\Services\DAL%emulation_extension%] >> %TempFile%
echo "APPLETSLOCALE"="%AppletsDir%" >> %TempFile%
echo. >> %TempFile%

echo [HKEY_LOCAL_MACHINE\SOFTWARE\Intel\Services\DAL%emulation_extension%] >> %TempFile%
echo "JHI_TRANSPORT_TYPE"=dword:0000000%Transport_Type% >> %TempFile%
echo. >> %TempFile%

echo [HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\services\eventlog\Application\IntelDalJhi%emulation_extension%] >> %TempFile%
echo "EventMessageFile"="%JHIDir%\\jhi_service.exe" >> %TempFile%
echo . >> %TempFile%

echo [HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\services\eventlog\Application\IntelDalJhi%emulation_extension%] >> %TempFile%
echo "TypesSupported"=dword:00000007 >> %TempFile%
goto:eof

:RemoveRegistryKeys
set TempRegistryFile=remove.reg
IF EXIST %TempRegistryFile% DEL %TempRegistryFile%

echo Windows Registry Editor Version 5.00 >> %TempRegistryFile%
echo. >> %TempRegistryFile%
echo [-HKEY_LOCAL_MACHINE\SOFTWARE\Intel\Services\DAL%emulation_extension%] >> %TempRegistryFile%
echo. >> %TempRegistryFile%
echo [-HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\services\eventlog\Application\IntelDalJhi%emulation_extension%] >> %TempRegistryFile%

REGEDIT /S %TempRegistryFile%

DEL %TempRegistryFile%

IF %ERRORLEVEL% EQU 0 EXIT /B 0
EXIT /B 1


:CreateJHIDirectories
@echo off
IF %is_X64_OS% EQU N (
	call MD "%DalPath_X86_OS%"
) else (
	call MD "%DalPath_X64_OS_PRIMERY%"
	call MD "%DalPath_X64_OS_SECONDERY%"
)

call MD "%AppletsFolder%"

EXIT /B 0

:RemoveJHIDirectories
@echo off
IF %is_X64_OS% EQU N (
	IF EXIST "%DalPath_X86_OS%" (
		call RD /S /Q "%DalPath_X86_OS%"
	)
) else (
	IF EXIST "%DalPath_X64_OS_PRIMERY%" (
		call RD /S /Q  "%DalPath_X64_OS_PRIMERY%"
	)
	IF EXIST "%DalPath_X64_OS_SECONDERY%" (
		call RD /S /Q  "%DalPath_X64_OS_SECONDERY%"
	)
)

IF EXIST "%AppletsFolder%" (
	call RD /S /Q "%AppletsFolder%"
)

EXIT /B 0

:CopyJHIFiles
@echo off
IF %is_X64_OS% EQU N (
	copy jhi_service.exe "%DalPath_X86_OS%\jhi_service.exe" > NUL
	copy JHI.dll "%DalPath_X86_OS%\JHI.dll" > NUL
	copy bhPlugin.dll "%DalPath_X86_OS%\bhPlugin.dll" > NUL
	copy bhPluginV2.dll "%DalPath_X86_OS%\bhPluginV2.dll" > NUL
	copy TeeManagement.dll "%DalPath_X86_OS%\TeeManagement.dll" > NUL
	copy SpoolerApplet.dalp "%DalPath_X86_OS%\SpoolerApplet.dalp" > NUL
	copy TEETransport.dll "%DalPath_X86_OS%\TEETransport.dll" > NUL
) else (
	copy jhi_service.exe "%DalPath_X64_OS_PRIMERY%\jhi_service.exe" > NUL
	copy JHI.dll "%DalPath_X64_OS_PRIMERY%\JHI.dll" > NUL
	copy bhPlugin.dll "%DalPath_X64_OS_PRIMERY%\bhPlugin.dll" > NUL
	copy bhPluginV2.dll "%DalPath_X64_OS_PRIMERY%\bhPluginV2.dll" > NUL
	copy TeeManagement.dll "%DalPath_X64_OS_PRIMERY%\TeeManagement.dll" > NUL
	copy SpoolerApplet.dalp "%DalPath_X64_OS_PRIMERY%\SpoolerApplet.dalp" > NUL
	copy TEETransport.dll "%DalPath_X64_OS_PRIMERY%\TEETransport.dll" > NUL
	
	copy JHI64.dll "%DalPath_X64_OS_SECONDERY%\JHI64.dll" > NUL
	copy TeeManagement64.dll "%DalPath_X64_OS_SECONDERY%\TeeManagement64.dll" > NUL
)
EXIT /B 0

:StopJHIService
@echo off 
set ExitCode=0
IF %is_X64_OS% EQU N (
	IF EXIST "%DalPath_X86_OS%\JHI_Service.exe" (
		call "%DalPath_X86_OS%\JHI_Service.exe" stop > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Stop JHI Service failed. JHI_Service.exe does not exist
	)
) else (
	IF EXIST "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" (
		call "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" stop > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Stop JHI Service failed. JHI_Service.exe does not exist
	)
)

REM - valid exit codes 0, 6
IF %ExitCode%==0 (
	echo JHI Service stopped.
	EXIT /B 0
) else IF %ExitCode%==6 (
	echo JHI Service stopped.
	EXIT /B 0
) else (
	echo Error: failed to stop JHI Service!
	EXIT /B 1
)
goto:eof

:UninstallJHIService
set ExitCode=0
IF %is_X64_OS% EQU N (
	IF EXIST "%DalPath_X86_OS%\JHI_Service.exe" (
		call "%DalPath_X86_OS%\JHI_Service.exe" uninstall > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Unregister JHI Service failed. JHI_Service.exe does not exist
	)
) else (
	IF EXIST "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" (
		call "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" uninstall > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Unregister JHI Service failed. JHI_Service.exe does not exist
	)
)

REM - valid exit codes 0, 4
IF %ExitCode%==0 (
	echo JHI Service uninstalled successfuly.
	EXIT /B 0
) else IF %ExitCode%==4 (
	echo JHI Service uninstalled successfuly.
	EXIT /B 0
) else (
	echo Error: failed to uninstall JHI Service!
	EXIT /B 1
)

goto:eof

:StartJHIService
set ExitCode=1
IF %is_X64_OS% EQU N (
	IF EXIST "%DalPath_X86_OS%\JHI_Service.exe" (
		call "%DalPath_X86_OS%\JHI_Service.exe" start > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Start JHI Failed. JHI_Service.exe does not exist
	)
) else (
	IF EXIST "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" (
		call "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" start > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Start JHI Failed. JHI_Service.exe does not exist
	)
)

REM - valid exit codes 0, 5
IF %ExitCode%==0 (
	echo JHI Service started.
	EXIT /B 0
) else IF %ExitCode%==5 (
	echo JHI Service started.
	EXIT /B 0
) else (
	echo Error: failed to start JHI Service!
	EXIT /B 1
)

goto:eof

:InstallJHIService
set ExitCode=1
IF %is_X64_OS% EQU N (
	IF EXIST "%DalPath_X86_OS%\JHI_Service.exe" (
		call "%DalPath_X86_OS%\JHI_Service.exe" install > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Register JHI Service failed. JHI_Service.exe does not exist
	)
) else (
	IF EXIST "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" (
		call "%DalPath_X64_OS_PRIMERY%\JHI_Service.exe" install > NUL
		set ExitCode=%ERRORLEVEL%
	) else (
		echo Register JHI Service failed. JHI_Service.exe does not exist
	)
)

IF %ExitCode%==0 (
	echo JHI Service installed successfuly.
	EXIT /B 0
) else IF %ExitCode%==3 (
	echo JHI Service installed successfuly.
	EXIT /B 0
) else (
	echo Error: failed to register JHI Service!
	EXIT /B 1
)
goto:eof


:VerifyJhiFilesExists
@echo off
set JHIFilesExist=Y

call:CHECK_FILE_EXIST JHI_Service.exe
call:CHECK_FILE_EXIST JHI.dll
call:CHECK_FILE_EXIST SpoolerApplet.dalp
call:CHECK_FILE_EXIST bhPlugin.dll
call:CHECK_FILE_EXIST bhPluginV2.dll
call:CHECK_FILE_EXIST TeeManagement.dll
call:CHECK_FILE_EXIST TEETransport.dll
IF %is_X64_OS% EQU Y call:CHECK_FILE_EXIST JHI64.dll
IF %is_X64_OS% EQU Y call:CHECK_FILE_EXIST TeeManagement64.dll

IF NOT %JHIFilesExist% EQU Y (
	EXIT /B 1
)
EXIT /B 0

:CHECK_FILE_EXIST
@echo off
if NOT EXIST %1% (
	echo Error: could not find %1%
	set JHIFilesExist=N
)
goto:eof

:VerifyAdminPrivileges
@echo off
NET SESSION > NUL
IF NOT %ERRORLEVEL% EQU 0 (
   echo.
   echo ######## ########  ########   #######  ########  
   echo ##       ##     ## ##     ## ##     ## ##     ## 
   echo ##       ##     ## ##     ## ##     ## ##     ## 
   echo ######   ########  ########  ##     ## ########  
   echo ##       ##   ##   ##   ##   ##     ## ##   ##   
   echo ##       ##    ##  ##    ##  ##     ## ##    ##  
   echo ######## ##     ## ##     ##  #######  ##     ## 
   echo.
   echo.
   echo ############### ERROR: ADMINISTRATOR PRIVILEGES REQUIRED ################
   echo #
   echo #  This script must be run as administrator to work properly!  
   echo #  If you're seeing this after clicking on the install.bat file,
   echo #  then right click on the file and select "Run As Administrator".
   echo #
   echo #########################################################################
   echo.
   PAUSE
   EXIT /B 1
)
goto:eof

:PrintLogo
@echo off
cls
echo.
echo. 
echo      ____. ___ ___ .___  .___                 __          .__   .__                  
echo     ^|   ^|^/   ^|   ^\^|   ^| ^|   ^| ____   _______^/  ^|______   ^|  ^|  ^|  ^|    ____ ___ 
echo     ^|   ^/    ~    ^\   ^| ^|   ^|^/    ^\ ^/  ^___^/^\   __^\__  ^\  ^|  ^|  ^|  ^|  _^/ __ ^\^\_  _ 
echo ^/^\__^|   ^\    Y    ^/   ^| ^|   ^|   ^|  ^\^\___ ^\  ^|  ^|  ^/ __ ^\_^|  ^|__^|  ^|__^\  ___^/ ^|  ^| ^\^/
echo ^\_______^|^\___^|_  ^/^|___^| ^|___^|___^|  ^/____  ^> ^|__^| (____  ^/^|____^/^|____^/ ^\___  ^>^|_
echo                ^\^/                ^\^/     ^\^/            ^\^/                  ^\^/        
echo                                                                  By Elad Dabool
echo.
goto:eof

:PrintUsage
@echo off
echo.
echo Usage: Installer.bat [command] [mode]
echo.
echo   Commands:
echo 	install    - install the JHI service using current directory files
echo 	uninstall  - remove the JHI service and all its resources
echo 	start      - starts the JHI service
echo 	stop       - stops the JHI service
echo.       
echo   Modes:
echo 	mei	   - applies the command to the JHI service (applying HECI registries)
echo 	sockets  - applies the command to the JHI service (applying sockets registries)
echo 	emulation  - applies the command to the JHI emulation service (for SDK) (applying sockets registries)
pause 
goto:eof



:AddFoldersToPath.program
	
	:: search for (x86)
	if ["%programFiles(x86)%"] == [""] goto :AddFoldersToPath.x64
	
	:: searching for x86 entry in the path...
	set foundInPath=false
	set jhiPath="%programFiles(x86)%\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%"
	call :AddFoldersToPath.parse "%path%"
	if %foundInPath%==true (
		rem echo %jhiPath% found in the PATH!
	) else (
		set pathToAddx86=";%programFiles(x86)%\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%"
	)
	
	
	:AddFoldersToPath.x64
	:: searching for x64 entry in the path...
	set foundInPath=false
	set jhiPath="%programFiles%\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%"
	call :AddFoldersToPath.parse "%path%"
	if %foundInPath%==true (
		rem echo %jhiPath% found in the PATH!
	) else (
		set pathToAddx64=";%programFiles%\Intel\Intel(R) Management Engine Components\DAL%emulation_extension%"
	)
	
	if [%pathToAddx86%] == [""] ( 
		if [%pathToAddx64%] == [""] (
			rem nothing to add
			exit /b 0 
		)
	)

	:: %programFiles(x86)% is defined
	call :AddFoldersToPath.setPath %pathToAddx64% %pathToAddx86%
	if not %errorlevel%==0 exit /b 1
	exit /b 0
::

:AddFoldersToPath.setPath
	setlocal
	call :AddFoldersToPath.strlen result "%PATH%%~1%~2"
	if 1024 LSS %result% (
		echo cannot write to path because it is bigger than 1024!
		echo path + new value length = %result%
		exit /b 1
	)
	::echo adding it to the path.
	setx /m path "%PATH%%~1%~2" >nul
	if not %errorlevel%==0 exit /b 1
	goto :eof
::

:AddFoldersToPath.parse
	set list=%1
	set list=%list:"=%

	FOR /f "tokens=1* delims=;" %%a IN ("%list%") DO (
		if "%%a" == %jhiPath% (
			set foundInPath=true
			goto :eof
		)
		if not "%%b" == "" call :AddFoldersToPath.parse "%%b"
	)

	goto :eof
::

:AddFoldersToPath.strlen <resultVar> <stringVar>
(   
    setlocal EnableDelayedExpansion
    set "s=!%~2!#"
    set "len=0"
    for %%P in (4096 2048 1024 512 256 128 64 32 16 8 4 2 1) do (
        if "!s:~%%P,1!" NEQ "" ( 
            set /a "len+=%%P"
            set "s=!s:~%%P!"
        )
    )
)
( 
    endlocal
    set "%~1=%len%"
    exit /b
)
::