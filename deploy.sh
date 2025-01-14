#!/bin/bash

# 配置
APP_NAME="api-order"
PORTS=(3133 3134)  # 定义两个端口
HEALTH_CHECK_PATH="/api/pay/health"
LOG_FILE="runtime.log"
BACKUP_DIR="backups"
SOURCE_DIR="api-pay"
GO_VERSION="go1.20.3"

# 日志函数
log() {
    local level=$1
    local message=$2
    local timestamp=$(date +'%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message"
    echo "[$timestamp] [$level] $message" >> $LOG_FILE
}

# 错误处理函数
handle_error() {
    local error_code=$?
    log "ERROR" "脚本在第 $1 行失败，退出代码 $error_code"
    exit $error_code
}

# 设置错误处理
trap 'handle_error ${LINENO}' ERR

# 检查和创建必要目录
init_directories() {
    if [ ! -d "$BACKUP_DIR" ]; then
        log "INFO" "创建备份目录: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
}

# 检查必要的工具
check_requirements() {
    for cmd in go git curl; do
        if ! command -v $cmd &> /dev/null; then
            log "ERROR" "$cmd 是必需的，但未安装"
            exit 1
        fi
    done
}

# 构建应用
build_application() {
    cd "$SOURCE_DIR" || exit

    log "INFO" "拉取最新代码..."
    git pull

    log "INFO" "检查依赖..."
    go mod tidy || { log "ERROR" "更新依赖失败"; exit 1; }

    log "INFO" "构建应用..."
    GOOS=linux GOARCH=amd64 go build -o "$APP_NAME" || { log "ERROR" "构建失败"; exit 1; }

    TMP_FILE="../${APP_NAME}.tmp"
    cp "$APP_NAME" "$TMP_FILE" || { log "ERROR" "临时文件复制失败"; exit 1; }

    log "INFO" "重新构建可执行文件成功！"

    cd ../

    cp -r api-pay/static/ ./

    mv "${APP_NAME}.tmp" "$APP_NAME" || { log "ERROR" "重命名文件失败"; exit 1; }
}

# 健康检查特定端口
check_health() {
    local port=$1
    local retry_count=0
    local max_retries=12
    local health_url="http://localhost:${port}${HEALTH_CHECK_PATH}"

    while [ $retry_count -lt $max_retries ]; do
        if curl -s $health_url | grep -q "ok"; then
            log "INFO" "端口 $port 健康检查通过"
            return 0
        fi

        retry_count=$((retry_count + 1))
        log "INFO" "等待端口 $port 服务变为健康... (尝试 $retry_count/$max_retries)"
        sleep 5
    done

    log "ERROR" "端口 $port 健康检查在 $max_retries 次尝试后失败"
    return 1
}

# 获取特定端口的进程ID
get_pid_by_port() {
    local port=$1
    lsof -i :$port -t
}

# 备份当前二进制
backup_binary() {
    local port=$1
    if [ -f "$APP_NAME" ]; then
        local timestamp=$(date +%Y%m%d_%H%M%S)
        cp $APP_NAME "$BACKUP_DIR/${APP_NAME}_${port}_${timestamp}"
        log "INFO" "备份端口 $port 的当前二进制"
    fi
}

# 启动新进程
start_new_process() {
    local port=$1
    log "INFO" "在端口 $port 启动新实例..."

    # 使用环境变量传递端口
    PORT=$port nohup ./$APP_NAME >> $LOG_FILE 2>&1 &
    NEW_PID=$!

    sleep 2

    if ps -p $NEW_PID > /dev/null; then
        log "INFO" "端口 $port 的进程成功启动，PID: $NEW_PID"
        return 0
    else
        log "ERROR" "端口 $port 的进程启动失败"
        return 1
    fi
}

# 停止特定端口的进程
stop_process() {
    local port=$1
    local pid=$(get_pid_by_port $port)

    if [ ! -z "$pid" ]; then
        log "INFO" "停止端口 $port 的进程 (PID: $pid)"
        kill -SIGTERM $pid
        sleep 2

        if ps -p $pid > /dev/null; then
            log "ERROR" "无法停止端口 $port 的进程"
            return 1
        fi
    fi
    return 0
}

# 更新特定端口的实例
update_instance() {
    local port=$1
    log "INFO" "开始更新端口 $port 的实例"

    # 备份当前实例
    backup_binary $port

    # 停止当前实例
    stop_process $port

    # 启动新实例
    if ! start_new_process $port; then
        log "ERROR" "端口 $port 启动新实例失败"
        return 1
    fi

    # 健康检查
    if ! check_health $port; then
        log "ERROR" "端口 $port 健康检查失败"
        return 1
    fi

    log "INFO" "端口 $port 更新完成"
    return 0
}

# 清理旧备份
cleanup_backups() {
    cd "$BACKUP_DIR" && ls -t | tail -n +6 | xargs -r rm
    log "INFO" "清理旧备份"
}

# 主函数
main() {
    init_directories
    check_requirements
    build_application

    # 依次更新每个端口
    for port in "${PORTS[@]}"; do
        # 分割线
        echo "------------------------"
        log "INFO" "开始处理端口 $port"

        if ! update_instance $port; then
            log "ERROR" "端口 $port 更新失败"
            # 这里可以添加回滚逻辑
            exit 1
        fi

        log "INFO" "等待确保服务稳定..."
        sleep 3
    done
    echo "------------------------"

    cleanup_backups
    log "INFO" "所有端口更新完成"
}

# 执行主函数
main