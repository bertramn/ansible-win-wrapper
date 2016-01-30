@echo off

REM If you used the stand Cygwin installer this will be C:\cygwin
rem set CYGWIN=%USERPROFILE%\.babun\cygwin
set CYGWIN=c:\cygwin
rem set CYGWIN=C:\dev\apps\MobaXterm\mbx-root
rem set CYGWIN=c:\dev\apps\babun\.babun\cygwin

REM You can switch this to work with bash with %CYGWIN%\bin\bash.exe
rem set SH=%CYGWIN%\bin\zsh.exe
set SH=%CYGWIN%\bin\bash.exe

rem cygwin
set EXEC_ANSIBLE_PLAYBOOK=/usr/bin/ansible-playbook
rem babun
rem set EXEC_ANSIBLE_PLAYBOOK=/bin/ansible-playbook

echo "----------------------------- " > c:\temp\vansible.log
setlocal enabledelayedexpansion
REM
REM set argCount=0
REM for %%x in (%*) do (
REM    set /A argCount+=1
REM    set "argVec[!argCount!]=%%~x"
REM )
REM
REM echo Number of processed arguments: %argCount% >> c:\temp\vansible.log
REM
REM for /L %%i in (1,1,%argCount%) do (
REM   echo %%i- !argVec[%%i]! >> c:\temp\vansible.log
REM   echo "Arg is %%i !argVec[%%i]:~0,14!"
REM   IF "!argVec[%%i]:~0,14!"=="--extra-vars='" (
REM      echo yes---------------
REM      echo !argVec[%%i]:~0,14!
REM   )
REM )

REM set CMD_LINE_ARGS=
REM set argCount=0
REM :setArgs
REM if ""%1""=="""" goto doneSetArgs
REM set /A argCount+=1
REM echo Add arg !argCount!: %1 >> c:\temp\vansible.log
REM set CMD_LINE_ARGS=%CMD_LINE_ARGS% %1
REM shift
REM goto setArgs
REM :doneSetArgs
REM echo "CMD: %CMD_LINE_ARGS%"

REM "%SH%" -c "%EXEC_ANSIBLE_PLAYBOOK% -vvvv %CMD_LINE_ARGS%"
"%SH%" -c "%EXEC_ANSIBLE_PLAYBOOK% %*"

rem c:\dev\apps\python\python-2.7.10\python.exe  C:\dev\apps\babun\.babun\cygwin\bin\ansible-playbook
