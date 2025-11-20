#!/usr/bin/env pwsh

# 批处理执行所有PowerShell脚本
$currentDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# 获取当前目录下所有.ps1脚本，排除build-all.ps1本身
$scripts = Get-ChildItem -Path $currentDir -Filter "*.ps1" | Where-Object { $_.Name -ne "build-all.ps1" }

# 依次执行每个脚本
foreach ($script in $scripts) {
    Write-Host "============================================="
    Write-Host "正在执行: $($script.Name)"
    Write-Host "============================================="
    
    try {
        & $script.FullName
        Write-Host "成功执行: $($script.Name)"
    } catch {
        Write-Host "执行失败: $($script.Name)"
        Write-Host "错误信息: $_"
    }
    
    Write-Host ""
}

Write-Host "============================================="
Write-Host "所有脚本执行完毕"
Write-Host "============================================="
