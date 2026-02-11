@echo off
REM Kube-Sentinel Mock Test Script
REM No Minikube needed - validates core code logic directly

echo ========================================
echo   Kube-Sentinel Mock Tests
echo   Testing Informer and Monitor Logic
echo ========================================
echo.

REM Setup Go environment
set GOROOT=D:\Go
set PATH=D:\Go\bin;%PATH%

cd /d D:\Coding\Kube\kube-sentinel

echo [1/3] Running config module tests...
go test -v ./internal/config
echo.

echo [2/3] Running Pod monitor module tests...
go test -v ./internal/monitor
echo.

echo [3/3] Running all tests with coverage...
go test -v -cover ./...
echo.

echo ========================================
echo   All Tests Completed!
echo ========================================
pause
