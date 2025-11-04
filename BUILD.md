# 构建说明

本文档说明如何构建输入法切换工具。

## 快速开始

使用提供的构建脚本可以快速构建项目：

```bash
# 基本构建
./build.sh

# 清理并构建
./build.sh --clean

# 构建并运行测试
./build.sh --test

# 完整构建（清理 + 测试）
./build.sh --clean --test
```

## 构建脚本功能

构建脚本 `build.sh` 提供以下功能：

- ✅ **环境检查**：自动检查 Go 环境和项目文件
- ✅ **依赖管理**：自动整理 Go 模块依赖
- ✅ **目录创建**：自动创建构建目录结构
- ✅ **并行构建**：生成可执行文件和 macOS 应用包
- ✅ **权限设置**：自动设置正确的文件权限
- ✅ **版本信息**：显示构建版本和时间信息
- ✅ **彩色输出**：提供清晰的状态反馈

## 构建产物

构建完成后会生成以下文件：

```
build/
├── bin/
│   ├── switch-input              # 可执行文件
│   ├── switch-input.app/         # macOS 应用包
│   │   └── Contents/
│   │       └── MacOS/
│   │           └── switch-input  # 应用包内可执行文件
│   └── logs/                     # 日志目录
```

## 运行方式

构建完成后有两种运行方式：

### 1. 直接运行可执行文件

```bash
./build/bin/switch-input
```

### 2. 使用 macOS 应用包

```bash
open build/bin/switch-input.app
```

## 环境变量

可以通过环境变量控制构建行为：

```bash
# 启用清理
CLEAN=true ./build.sh

# 启用测试
RUN_TESTS=true ./build.sh

# 同时启用清理和测试
CLEAN=true RUN_TESTS=true ./build.sh
```

## 手动构建

如果需要手动构建，可以执行以下命令：

```bash
# 1. 整理依赖
go mod tidy

# 2. 创建构建目录
mkdir -p build/bin
mkdir -p build/bin/switch-input.app/Contents/MacOS

# 3. 构建项目
go build -o build/bin/switch-input

# 4. 复制到应用包
cp build/bin/switch-input build/bin/switch-input.app/Contents/MacOS/
chmod +x build/bin/switch-input.app/Contents/MacOS/switch-input
```

## 系统要求

- **Go**: 1.23 或更高版本
- **macOS**: 支持 Intel 和 Apple Silicon
- **im-select**: 需要安装在 `/opt/homebrew/bin/im-select`

## 故障排除

### 构建失败

1. 检查 Go 版本：`go version`
2. 检查项目文件：确认 `go.mod` 和 `main.go` 存在
3. 清理后重新构建：`./build.sh --clean`

### 运行失败

1. 确认 `im-select` 工具已正确安装
2. 检查文件权限：`ls -la build/bin/switch-input`
3. 查看日志输出：程序运行时会输出详细信息

## 开发提示

- 修改代码后，只需运行 `./build.sh` 即可快速重新构建
- 使用 `--test` 选项可以在构建后自动运行测试
- 构建脚本会自动显示文件大小和版本信息
- 所有构建步骤都有详细的状态输出和错误处理