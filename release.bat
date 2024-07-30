@echo off
setlocal enabledelayedexpansion

REM 设置代理（如果需要）
set http_proxy=http://spiritlhl:spiritlhl@202.81.231.25:44516
set https_proxy=http://spiritlhl:spiritlhl@202.81.231.25:44516

REM 设置仓库路径
set repo_path=C:\Users\spiritlhl\Documents\GoWorks\ecs

REM 进入仓库目录
cd %repo_path%

REM 添加所有更改并提交
git add -A
git commit -am "update"

REM 推送代码到 master 分支并创建标签
:push
git -c http.proxy=%http_proxy% -c https.proxy=%https_proxy% -c http.sslVerify=false -c https.sslVerify=false push -f origin master
if errorlevel 1 (
    echo Push failed. Retrying in 3 seconds...
    timeout /nobreak /t 3 >nul
    goto push
)

REM 提示用户输入版本号
set /p version="Enter the version number (e.g., v1.0.0): "

REM 创建并推送标签
:push_tag
git -c http.proxy=%http_proxy% -c https.proxy=%https_proxy% -c http.sslVerify=false -c https.sslVerify=false tag %version%
git -c http.proxy=%http_proxy% -c https.proxy=%https_proxy% -c http.sslVerify=false -c https.sslVerify=false push origin %version%
if errorlevel 1 (
    echo Tag push failed. Retrying in 3 seconds...
    timeout /nobreak /t 3 >nul
    goto push_tag
)

echo Push and tag creation successful.

endlocal
