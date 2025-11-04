#!/bin/bash

# 输入法切换工具构建脚本
# 自动化构建、更新可执行文件和应用包

set -e  # 遇到错误立即退出

# 颜色输出定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Go 是否安装
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go: https://golang.org/dl/"
        exit 1
    fi
    log_success "Go 版本: $(go version)"
}

# 检查项目目录
check_project() {
    if [[ ! -f "go.mod" ]]; then
        log_error "这不是一个 Go 项目目录 (缺少 go.mod)"
        exit 1
    fi

    if [[ ! -f "main.go" ]]; then
        log_error "找不到 main.go 文件"
        exit 1
    fi
}

# 创建构建目录
create_build_dirs() {
    log_info "创建构建目录..."
    mkdir -p build/bin
    mkdir -p build/bin/switch-input.app/Contents/MacOS
    mkdir -p build/bin/logs
    log_success "构建目录已创建"
}

# 整理依赖
tidy_deps() {
    log_info "整理 Go 模块依赖..."
    go mod tidy
    log_success "依赖整理完成"
}

# 构建项目
build_project() {
    log_info "开始构建项目..."

    # 获取版本信息
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
    GO_VERSION=$(go version | awk '{print $3}')

    # 构建参数 (如果 main 包中没有这些变量，则不使用 ldflags)
    # LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GoVersion=${GO_VERSION}"

    # 执行构建
    go build -o build/bin/switch-input

    if [[ $? -eq 0 ]]; then
        log_success "项目构建成功"
        log_info "构建信息:"
        log_info "  - 版本: ${VERSION}"
        log_info "  - 构建时间: ${BUILD_TIME}"
        log_info "  - Go 版本: ${GO_VERSION}"
    else
        log_error "项目构建失败"
        exit 1
    fi
}

# 检查构建结果
check_build() {
    if [[ ! -f "build/bin/switch-input" ]]; then
        log_error "构建失败：找不到可执行文件"
        exit 1
    fi

    # 获取文件大小
    FILE_SIZE=$(ls -lh build/bin/switch-input | awk '{print $5}')
    log_success "可执行文件已生成: build/bin/switch-input (${FILE_SIZE})"
}

# 更新 macOS 应用包
update_app_bundle() {
    log_info "更新 macOS 应用包..."

    # 复制可执行文件到应用包
    cp build/bin/switch-input build/bin/switch-input.app/Contents/MacOS/

    # 设置执行权限
    chmod +x build/bin/switch-input.app/Contents/MacOS/switch-input

    log_success "macOS 应用包已更新: build/bin/switch-input.app"
}

# 运行测试 (可选)
run_tests() {
    if [[ -n "$RUN_TESTS" ]] && [[ "$RUN_TESTS" == "true" ]]; then
        log_info "运行测试..."
        if go test ./...; then
            log_success "所有测试通过"
        else
            log_warning "测试失败，但继续构建"
        fi
    fi
}

# 清理旧文件 (可选)
clean_old_files() {
    if [[ -n "$CLEAN" ]] && [[ "$CLEAN" == "true" ]]; then
        log_info "清理旧文件..."
        rm -rf build/bin/switch-input
        rm -rf build/bin/switch-input.app/Contents/MacOS/switch-input
        log_success "清理完成"
    fi
}

# 显示使用帮助
show_help() {
    echo "输入法切换工具构建脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -c, --clean    构建前清理旧文件"
    echo "  -t, --test     构建后运行测试"
    echo "  -v, --verbose  显示详细输出"
    echo ""
    echo "环境变量:"
    echo "  CLEAN=true     清理旧文件"
    echo "  RUN_TESTS=true 运行测试"
    echo ""
    echo "示例:"
    echo "  $0                    # 基本构建"
    echo "  $0 --clean --test     # 清理并测试构建"
    echo "  CLEAN=true $0         # 使用环境变量清理"
}

# 解析命令行参数
VERBOSE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -t|--test)
            RUN_TESTS=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 主构建流程
main() {
    log_info "开始构建输入法切换工具..."
    log_info "项目目录: $(pwd)"

    # 检查环境
    check_go
    check_project

    # 清理 (如果需要)
    if [[ -n "$CLEAN" ]] && [[ "$CLEAN" == "true" ]]; then
        clean_old_files
    fi

    # 准备构建环境
    create_build_dirs

    # 依赖管理
    tidy_deps

    # 运行测试 (如果需要)
    if [[ -n "$RUN_TESTS" ]] && [[ "$RUN_TESTS" == "true" ]]; then
        run_tests
    fi

    # 构建项目
    build_project
    check_build

    # 更新应用包
    update_app_bundle

    # 显示最终结果
    echo ""
    log_success "🎉 构建完成！"
    echo ""
    echo "构建产物:"
    echo "  📦 可执行文件: build/bin/switch-input"
    echo "  📱 macOS 应用: build/bin/switch-input.app"
    echo ""
    echo "运行方式:"
    echo "  直接运行: ./build/bin/switch-input"
    echo "  应用包: open build/bin/switch-input.app"
    echo ""
}

# 执行主函数
main