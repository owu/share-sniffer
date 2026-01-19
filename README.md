# Share Sniffer (分享嗅探器)

[English](./README_en.md) | 中文 | [日本語](./README_jp.md)


## 1、介绍

Share Sniffer（分享嗅探器）是一款跨平台的网盘分享链接检测工具，支持多种主流网盘的分享链接有效性检测。该工具提供了直观的图形界面（GUI）和便捷的命令行界面（CLI），方便用户根据需求选择使用方式。

### 1.1 支持的网盘类型

- ✅ 夸克网盘
- ✅ 天翼云盘
- ✅ 百度网盘
- ✅ 阿里云盘
- ✅ 115网盘
- ✅ 123云盘
- ✅ UC网盘
- ✅ 迅雷云盘
- ✅ 移动云盘

## 2、起源

某影视资源分享群有一份在线表格，包含数几千个影视资源分享链接，分享链接时有过期人工检查慢，因此有了这个工具。通过自动化检测，可以快速筛选出有效的分享链接，提高资源管理效率。

## 3、技术栈

- **开发语言**：Go 1.25
- **GUI框架**：[fyne.io/fyne/v2](https://fyne.io/) - 跨平台GUI框架
- **CLI框架**：[github.com/spf13/cobra](https://github.com/spf13/cobra) - 命令行框架


## 4、开发运行

### 4.1 GUI模式

#### 4.1.1 直接运行

```bash
# 初始化依赖
go mod tidy

# 运行GUI应用
go run ./launcher/gui/main.go

# 运行GUI应用，同时打印出编译和链接过程中执行的所有详细命令
go clean -cache  && go clean -modcache && go run -x ./launcher/gui/main.go

```

#### 4.1.2 开发模式

```bash
# 安装 fyne 开发工具（可选）
go install fyne.io/tools/cmd/fyne@latest

# 使用 fyne 开发模式运行（支持热重载）
fyne serve -src ./launcher/gui
```

### 4.2 CLI模式

#### 4.2.1 直接运行

```bash
# 运行CLI应用
go run ./launcher/cli/main.go [命令/URL]
```

#### 4.2.2 CLI命令示例

```bash
# 显示帮助信息
./share-sniffer-cli --help

# 查看版本
./share-sniffer-cli version

# 查看支持的链接类型
./share-sniffer-cli support

# 查看项目主页
./share-sniffer-cli home

# 检测单个链接
./share-sniffer-cli "https://pan.quark.cn/s/0a6e84c02020"
```

## 5、打包编译

项目提供了自动化打包脚本，位于 `/build` 目录下，支持 Windows 和 Linux 平台的打包。

### 5.1 打包脚本说明

| 脚本名称 | 平台 | 说明 |
|---------|------|------|
| `build-gui-windows.ps1` | Windows | PowerShell脚本，用于构建Windows平台的GUI可执行文件 |
| `build-gui-linux.sh` | Linux | Bash脚本，用于构建Linux平台的GUI安装包 |
| `build-android.ps1` | Windows | PowerShell脚本，用于构建Android平台的APK |
| `build-android.sh` | Linux | Bash脚本，用于构建Android平台的APK |
| `build-cli-windows.ps1` | Windows | PowerShell脚本，用于构建Windows平台的CLI可执行文件 |
| `build-cli-linux.sh` | Linux | Bash脚本，用于构建Linux平台的CLI可执行文件 |
| `build-all.ps1` | Windows | PowerShell脚本，用于批量构建Windows平台的所有可执行文件 |
| `build-all.sh` | Linux | Bash脚本，用于批量构建Linux平台的所有可执行文件 |

### 5.2 使用打包脚本

#### 5.2.1 Windows平台

```powershell
# 构建Windows GUI版本
cd build/scripts
./build-gui-windows.ps1

# 构建Android版本
cd build/scripts
./build-android.ps1

# 构建CLI工具
cd build/scripts
./build-cli-windows.ps1

# 批量构建所有Windows包
cd build/scripts
./build-all.ps1
```

#### 5.2.2 Linux平台

```bash
# 构建Linux GUI版本
cd build/scripts
chmod +x *.sh
./build-gui-linux.sh

# 构建Android版本
./build-android.sh

# 构建CLI工具
./build-cli-linux.sh

# 批量构建所有Linux包
./build-all.sh
```

### 5.3 打包脚本特性

- 自动从 `internal/config/config.go` 中读取版本号
- 自动检测并安装 `fyne` 工具（如果未安装）
- 清理Go缓存，确保构建环境干净
- 生成的文件自动命名并输出到 `/build/releases/{version}/` 目录
- 支持Windows、Linux和Android平台
- 提供批量构建脚本，实现一键编译

## 6、安装及卸载

### 6.1 Linux GUI安装

1. 下载最新安装包 `ShareSniffer.v0.2.0.linux-amd64.tar.xz` 到任意目录

2. 文件解压，进入目录，安装：

```bash
# 创建安装目录
mkdir ./ShareSniffer.linux-amd64 

# 解压安装包
tar -xJf ./ShareSniffer.v0.2.0.linux-amd64.tar.xz -C ./ShareSniffer.linux-amd64 

# 进入安装目录
cd ./ShareSniffer.linux-amd64 

# 执行安装
sudo make install
```

### 6.2 Linux GUI卸载

```bash
# 进入安装目录
cd ./ShareSniffer.linux-amd64 

# 执行卸载
sudo make uninstall 

# 返回上级目录
cd ../ 

# 删除安装目录
rm -rf ./ShareSniffer.linux-amd64
```
### 6.3 share-sniffer-cli 安装
#### 6.3.1 Linux 安装
```
下载最新版 share-sniffer-cli.v0.2.0.linux-amd64
重命名为 share-sniffer-cli 
可执行文件移动到 `/usr/local/bin` 目录
```
#### 6.3.2 Windows 安装
```
下载最新版 share-sniffer-cli.v0.2.0.windows-amd64.exe
重命名为 share-sniffer-cli.exe
可执行文件移动到 `C:\Windows\System32` 目录（可选）
```


## 7、界面预览

### 7.1 关于界面

![about](https://github.com/owu/share-sniffer/raw/unstable/screenshot/about.png)

![update](https://github.com/owu/share-sniffer/raw/unstable/screenshot/update.png)

### 7.2 检测界面

![check](https://github.com/owu/share-sniffer/raw/unstable/screenshot/check.png)

![open](https://github.com/owu/share-sniffer/raw/unstable/screenshot/open.png)

![load](https://github.com/owu/share-sniffer/raw/unstable/screenshot/load.png)

![checking](https://github.com/owu/share-sniffer/raw/unstable/screenshot/checking.png)

### 7.3 结果界面

![result](https://github.com/owu/share-sniffer/raw/unstable/screenshot/result.png)

## 8、CLI模式工具介绍

### 8.1 命令说明

| 命令 | 说明 | 示例 |
|------|------|------|
| `help` | 显示帮助信息 | `./share-sniffer-cli --help` |
| `version` | 显示版本信息 | `./share-sniffer-cli version` |
| `support` | 显示支持的链接类型 | `./share-sniffer-cli support` |
| `home` | 显示项目主页链接 | `./share-sniffer-cli home` |
| `[URL]` | 检测指定链接 | `./share-sniffer-cli "https://pan.quark.cn/s/0a6e84c02020"` |

### 8.2 输出格式

CLI工具返回JSON格式的结果，方便其他程序调用：

```json
{
  "error": 0,
  "msg": "valid",
  "data": {
    "url": "https://pan.quark.cn/s/0a6e84c02020",
    "name": "国语动漫",
    "elapsed": 359
  }
}
```

#### 8.2.1 输出字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `error` | int | 错误码，0 表示 没有错误的，即链接有效；10 表示 未知错误；11 表示 链接过期的；12 表示 参数错误等；13 表示 超时的；14 表示 请求过程报错 |
| `msg` | string | 状态描述 |
| `data` | object | 检测结果详情 |
| `data.url` | string | 检测的URL |
| `data.name` | string | 资源名称（如果检测成功） |
| `data.elapsed` | int64 | 检测耗时（毫秒） |

### 8.3 使用场景

- 批量检测链接有效性
- 集成到其他脚本或程序中
- 服务器环境下使用
- 自动化检测工作流

## 9、贡献

欢迎提交 Issue 和 Pull Request！

## 10、许可证

[GNU GPL v3 License](LICENSE)
