@echo off
chcp 65001 >nul
title 轻量级诊所管理系统

echo.
echo ========================================
echo    轻量级诊所管理系统
echo ========================================
echo.

echo 正在启动服务器...
echo.

REM 检查可执行文件是否存在
if not exist "lighthospital.exe" (
    echo 错误: 找不到 lighthospital.exe 文件
    echo 请确保此批处理文件与 lighthospital.exe 在同一目录下
    pause
    exit /b 1
)

REM 检查模板文件是否存在
if not exist "templates" (
    echo 错误: 找不到 templates 目录
    echo 请确保此批处理文件与 templates 目录在同一目录下
    pause
    exit /b 1
)

REM 检查静态文件是否存在
if not exist "static" (
    echo 错误: 找不到 static 目录
    echo 请确保此批处理文件与 static 目录在同一目录下
    pause
    exit /b 1
)

echo 系统信息:
echo - 本地访问: http://localhost:8080
echo - 局域网访问: http://[本机IP]:8080
echo.
echo 默认登录信息:
echo - 管理员: admin / admin123
echo - 医生: doctor / doctor123
echo.

echo 按 Ctrl+C 停止服务器
echo.

REM 启动应用程序
lighthospital.exe

echo.
echo 服务器已停止
pause 