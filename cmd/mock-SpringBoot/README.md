# MemoryLogMonitor Mock Spring Boot 应用

这是一个基于 Spring Boot 2.7.9 的测试应用，用于测试向 MemoryLogMonitor 服务发送日志。

## 功能特性

- 简单的 Web 界面，用于输入和发送日志
- 支持选择日志级别（DEBUG/INFO/WARN/ERROR）
- 通过 Log4j2 SocketAppender 将日志发送到 MemoryLogMonitor（端口 9090）
- 自动格式化日志消息（包含时间戳和日志级别）

## 前置要求

1. Java 8 或更高版本
2. Maven 3.6+ 或 Gradle 6.9+
3. MemoryLogMonitor 服务已启动并监听 9090 端口

## 快速开始

### 1. 编译项目

```bash
cd cmd/mock-SpringBoot
mvn clean package
```

### 2. 运行应用

```bash
# 使用 Maven
mvn spring-boot:run

# 或使用 Java
java -jar target/mock-springboot-1.0.0.jar
```

### 3. 访问应用

打开浏览器访问：http://localhost:8081

### 4. 测试发送日志

1. 在页面上选择日志级别
2. 输入日志内容
3. 点击"发送日志"按钮
4. 在 MemoryLogMonitor Web 界面（http://localhost:8080）查看接收到的日志

## 配置说明

### 应用端口

默认端口为 `8081`，可在 `application.properties` 中修改：

```properties
server.port=8081
```

### MemoryLogMonitor 地址

默认连接到 `localhost:9090`，可通过环境变量修改：

```bash
# Linux/macOS
export LOG_HOST=192.168.1.100
export LOG_PORT=9090
mvn spring-boot:run

# Windows
set LOG_HOST=192.168.1.100
set LOG_PORT=9090
mvn spring-boot:run
```

### Log4j2 配置

Log4j2 配置文件位于 `src/main/resources/log4j2.xml`，包含：

- 控制台输出
- TCP Socket Appender（发送到 MemoryLogMonitor）
- 连接超时：1 秒
- 重连间隔：30 秒
- 忽略异常：true（避免影响主程序）

## 项目结构

```
cmd/mock-SpringBoot/
├── pom.xml                                    # Maven 配置文件
├── README.md                                  # 本文件
└── src/
    └── main/
        ├── java/
        │   └── com/
        │       └── memorylogmonitor/
        │           └── mock/
        │               ├── MockApplication.java          # 主应用类
        │               └── controller/
        │                   └── LogController.java         # 日志控制器
        └── resources/
            ├── application.properties                    # Spring Boot 配置
            ├── log4j2.xml                               # Log4j2 配置
            └── templates/
                └── index.html                           # 前端页面
```

## 故障排查

### 1. 日志未发送到 MemoryLogMonitor

- 检查 MemoryLogMonitor 服务是否启动
- 检查端口 9090 是否被占用
- 检查防火墙设置
- 查看应用控制台日志，确认是否有连接错误

### 2. 页面无法访问

- 检查应用是否成功启动
- 检查端口 8081 是否被占用
- 查看应用启动日志

### 3. 连接超时

- 确认 MemoryLogMonitor 服务地址和端口正确
- 检查网络连接
- 查看 Log4j2 配置中的超时设置

## 开发说明

### 修改日志格式

编辑 `src/main/resources/log4j2.xml` 中的 `LOG_PATTERN` 属性。

### 修改页面样式

编辑 `src/main/resources/templates/index.html` 中的 CSS 样式。

### 添加新功能

- Controller: `src/main/java/com/memorylogmonitor/mock/controller/LogController.java`
- 页面模板: `src/main/resources/templates/index.html`

## 参考

- [Spring Boot 2.7.9 文档](https://docs.spring.io/spring-boot/docs/2.7.9/reference/html/)
- [Log4j2 文档](https://logging.apache.org/log4j/2.x/manual/appenders.html#SocketAppender)
- [MemoryLogMonitor README](../README.md)
- [Spring Boot 集成指南](../../README-Java.md)
