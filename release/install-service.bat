@echo off
chcp 65001 >nul
title 轻量级诊所管理系统 - 服务安装

echo.
echo ========================================
echo    轻量级诊所管理系统 - 服务安装
echo ========================================
echo.

REM 检查管理员权限
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 需要管理员权限来安装服务
    echo 请右键点击此文件，选择"以管理员身份运行"
    pause
    exit /b 1
)

REM 检查 NSSM 是否存在
if not exist "nssm.exe" (
    echo 错误: 找不到 nssm.exe 文件
    echo 请确保 nssm.exe 与此批处理文件在同一目录下
    echo 您可以从 https://nssm.cc/download 下载 NSSM
    pause
    exit /b 1
)

REM 检查主程序是否存在
if not exist "lighthospital.exe" (
    echo 错误: 找不到 lighthospital.exe 文件
    pause
    exit /b 1
)

echo 正在安装 Windows 服务...
echo.

REM 获取当前目录的绝对路径
set "CURRENT_DIR=%~dp0"
set "CURRENT_DIR=%CURRENT_DIR:~0,-1%"

REM 安装服务
echo 安装服务: LightHospital
nssm.exe install LightHospital "%CURRENT_DIR%\lighthospital.exe"
if %errorLevel% neq 0 (
    echo 服务安装失败
    pause
    exit /b 1
)

REM 设置服务描述
nssm.exe set LightHospital Description "轻量级诊所管理系统 - 基于Go和Gin的Web应用"

REM 设置启动目录
nssm.exe set LightHospital AppDirectory "%CURRENT_DIR%"

REM 设置启动类型为自动
nssm.exe set LightHospital Start SERVICE_AUTO_START

REM 设置服务依赖
nssm.exe set LightHospital DependOnService Tcpip

REM 设置失败重启
nssm.exe set LightHospital AppRestartDelay 10000
nssm.exe set LightHospital AppStopMethodSkip 0
nssm.exe set LightHospital AppStopMethodConsole 1500
nssm.exe set LightHospital AppStopMethodWindow 1500
nssm.exe set LightHospital AppStopMethodThreads 1500

echo.
echo 服务安装完成！
echo.
echo 服务信息:
echo - 服务名称: LightHospital
echo - 启动类型: 自动
echo - 访问地址: http://localhost:8080
echo.
echo 管理命令:
echo - 启动服务: net start LightHospital
echo - 停止服务: net stop LightHospital
echo - 删除服务: nssm.exe remove LightHospital confirm
echo.

REM 询问是否立即启动服务
set /p "START_SERVICE=是否立即启动服务？(y/n): "
if /i "%START_SERVICE%"=="y" (
    echo 正在启动服务...
    net start LightHospital
    if %errorLevel% equ 0 (
        echo 服务启动成功！
        echo 您现在可以通过 http://localhost:8080 访问系统
    ) else (
        echo 服务启动失败，请检查日志
    )
)

echo.
echo 安装完成！
pause 