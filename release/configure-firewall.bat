@echo off
chcp 65001 >nul
title 轻量级诊所管理系统 - 防火墙配置

echo.
echo ========================================
echo    轻量级诊所管理系统 - 防火墙配置
echo ========================================
echo.

REM 检查管理员权限
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 需要管理员权限来配置防火墙
    echo 请右键点击此文件，选择"以管理员身份运行"
    pause
    exit /b 1
)

echo 正在配置 Windows 防火墙规则...
echo.

REM 添加入站规则
echo 添加入站规则...
netsh advfirewall firewall add rule name="LightHospital-Inbound" dir=in action=allow protocol=TCP localport=8080 description="轻量级诊所管理系统 - 入站规则"

REM 添加出站规则
echo 添加出站规则...
netsh advfirewall firewall add rule name="LightHospital-Outbound" dir=out action=allow protocol=TCP localport=8080 description="轻量级诊所管理系统 - 出站规则"

echo.
echo 防火墙配置完成！
echo.
echo 配置信息:
echo - 端口: 8080
echo - 协议: TCP
echo - 方向: 入站和出站
echo.
echo 现在其他设备可以通过以下地址访问:
echo - http://[本机IP]:8080
echo.
echo 要查看本机IP地址，请运行: ipconfig
echo.

pause 