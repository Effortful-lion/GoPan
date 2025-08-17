@echo off
echo go-mysql-transfer starting...
echo.

:: 切换到D盘
D:

:: 进入程序所在目录
cd D:\\GoLanddev\\go-mysql-transfer-v1.0.4\\transfer

:: 启动程序（假设程序文件名为go-mysql-transfer.exe）
echo running...
go-mysql-transfer.exe -config app.yml

:: 程序退出后暂停，方便查看日志
echo.
echo over, press any key to exit...
pause >nul
