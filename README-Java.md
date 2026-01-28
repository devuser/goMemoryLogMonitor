# Spring Boot 集成 MemoryLogMonitor

本文档说明如何在 Spring Boot 2.7.9 项目中使用 Log4j2 将日志输出到 MemoryLogMonitor 的 TCP 端口（9090）。

## 前置要求

1. Spring Boot 2.7.9 项目
2. Java 8 或更高版本
3. Log4j2 日志框架（Spring Boot 2.7.9 内置支持）
4. MemoryLogMonitor 服务已启动并监听 9090 端口
5. Maven 3.6+ 或 Gradle 6.9+

## 依赖配置

### Maven (Spring Boot 2.7.9)

在 `pom.xml` 中添加 Log4j2 依赖：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>2.7.9</version>
        <relativePath/>
    </parent>
    
    <groupId>com.yourcompany</groupId>
    <artifactId>your-application</artifactId>
    <version>1.0.0</version>
    
    <properties>
        <java.version>1.8</java.version>
    </properties>
    
    <dependencies>
        <!-- Spring Boot Starter -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter</artifactId>
            <exclusions>
                <!-- 排除 Spring Boot 默认的 Logback -->
                <exclusion>
                    <groupId>org.springframework.boot</groupId>
                    <artifactId>spring-boot-starter-logging</artifactId>
                </exclusion>
            </exclusions>
        </dependency>
        
        <!-- 添加 Log4j2 -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-log4j2</artifactId>
            <version>2.7.9</version>
        </dependency>
        
        <!-- Web 支持（如果需要） -->
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-web</artifactId>
            <exclusions>
                <exclusion>
                    <groupId>org.springframework.boot</groupId>
                    <artifactId>spring-boot-starter-logging</artifactId>
                </exclusion>
            </exclusions>
        </dependency>
    </dependencies>
    
    <build>
        <plugins>
            <plugin>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-maven-plugin</artifactId>
            </plugin>
        </plugins>
    </build>
</project>
```

### Gradle (Spring Boot 2.7.9)

在 `build.gradle` 中添加：

```gradle
plugins {
    id 'org.springframework.boot' version '2.7.9'
    id 'io.spring.dependency-management' version '1.0.15.RELEASE'
    id 'java'
}

group = 'com.yourcompany'
version = '1.0.0'
sourceCompatibility = '1.8'

repositories {
    mavenCentral()
}

configurations {
    all {
        exclude group: 'org.springframework.boot', module: 'spring-boot-starter-logging'
    }
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter'
    implementation 'org.springframework.boot:spring-boot-starter-log4j2'
    
    // Web 支持（如果需要）
    implementation('org.springframework.boot:spring-boot-starter-web') {
        exclude group: 'org.springframework.boot', module: 'spring-boot-starter-logging'
    }
}

tasks.named('test') {
    useJUnitPlatform()
}
```

### 依赖版本说明 (Spring Boot 2.7.9)

Spring Boot 2.7.9 会自动管理以下依赖版本：

- `spring-boot-starter-log4j2`: 2.7.9
- `log4j-core`: 2.17.2
- `log4j-api`: 2.17.2
- `log4j-slf4j-impl`: 2.17.2
- `log4j-jul`: 2.17.2

如果需要查看完整的依赖树：

```bash
# Maven
mvn dependency:tree

# Gradle
./gradlew dependencies
```

## Log4j2 配置

### 方式一：使用 SocketAppender（推荐）

创建或修改 `src/main/resources/log4j2.xml`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN">
    <Appenders>
        <!-- 控制台输出 -->
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
        </Console>
        
        <!-- TCP Socket Appender - 输出到 MemoryLogMonitor -->
        <Socket name="MemoryLogMonitor" host="localhost" port="9090" protocol="TCP">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
            <!-- 忽略异常，避免日志发送失败影响主程序 -->
            <IgnoreExceptions>true</IgnoreExceptions>
            <!-- 连接失败时重试间隔（毫秒） -->
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <!-- 立即刷新 -->
            <ImmediateFlush>true</ImmediateFlush>
            <!-- 连接超时（毫秒） -->
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </Socket>
    </Appenders>
    
    <Loggers>
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

### 方式二：使用 TcpSocketAppender（更灵活）

如果需要更细粒度的控制，可以使用 `TcpSocketAppender`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN">
    <Appenders>
        <!-- 控制台输出 -->
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
        </Console>
        
        <!-- TCP Socket Appender - 输出到 MemoryLogMonitor -->
        <TcpSocket name="MemoryLogMonitor" host="localhost" port="9090" protocol="TCP">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
            <!-- 忽略异常 -->
            <IgnoreExceptions>true</IgnoreExceptions>
            <!-- 连接失败时重试间隔（毫秒） -->
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <!-- 立即刷新 -->
            <ImmediateFlush>true</ImmediateFlush>
            <!-- 连接超时（毫秒） -->
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </TcpSocket>
    </Appenders>
    
    <Loggers>
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

### 方式三：使用 AsyncAppender（异步发送，性能更好）

如果日志量较大，建议使用异步方式发送：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN">
    <Appenders>
        <!-- 控制台输出 -->
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
        </Console>
        
        <!-- TCP Socket Appender -->
        <Socket name="MemoryLogMonitorSocket" host="localhost" port="9090" protocol="TCP">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
            <IgnoreExceptions>true</IgnoreExceptions>
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <ImmediateFlush>true</ImmediateFlush>
            <!-- 连接超时（毫秒） -->
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </Socket>
        
        <!-- 异步包装器 -->
        <Async name="MemoryLogMonitor" bufferSize="1024">
            <AppenderRef ref="MemoryLogMonitorSocket"/>
            <!-- 忽略异常 -->
            <IgnoreExceptions>true</IgnoreExceptions>
        </Async>
    </Appenders>
    
    <Loggers>
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

## 配置说明

### 关键参数

- **host**: MemoryLogMonitor 服务地址，默认 `localhost`
- **port**: MemoryLogMonitor TCP 端口，默认 `9090`
- **protocol**: 协议类型，使用 `TCP`
- **IgnoreExceptions**: 设置为 `true`，忽略发送失败异常，避免影响主程序
- **ReconnectDelayMillis**: 连接失败时重试间隔（毫秒），默认 30000（30秒）
- **ImmediateFlush**: 立即刷新，确保日志及时发送
- **ConnectTimeoutMillis**: 连接超时时间（毫秒），建议设置为 1000（1秒）

### 日志格式

默认日志格式：`%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n`

- `%d{yyyy-MM-dd HH:mm:ss.SSS}`: 时间戳
- `[%t]`: 线程名
- `%-5level`: 日志级别（左对齐，5个字符）
- `%logger{36}`: Logger 名称（最多36个字符）
- `%msg`: 日志消息
- `%n`: 换行符

可以根据需要自定义格式。

## 环境变量配置

为了支持不同环境，可以使用环境变量或系统属性：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN">
    <Properties>
        <!-- 从环境变量读取，默认 localhost:9090 -->
        <Property name="log.host">${env:LOG_HOST:-localhost}</Property>
        <Property name="log.port">${env:LOG_PORT:-9090}</Property>
    </Properties>
    
    <Appenders>
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
        </Console>
        
        <Socket name="MemoryLogMonitor" 
                host="${log.host}" 
                port="${log.port}" 
                protocol="TCP">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
            <IgnoreExceptions>true</IgnoreExceptions>
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <ImmediateFlush>true</ImmediateFlush>
            <!-- 连接超时（毫秒） -->
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </Socket>
    </Appenders>
    
    <Loggers>
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

启动时设置环境变量：

```bash
# Linux/macOS
export LOG_HOST=192.168.1.100
export LOG_PORT=9090
java -jar your-app.jar

# Windows
set LOG_HOST=192.168.1.100
set LOG_PORT=9090
java -jar your-app.jar

# 或者在启动命令中直接指定
java -DLOG_HOST=192.168.1.100 -DLOG_PORT=9090 -jar your-app.jar
```

### 下载依赖（Maven）

如果使用 Maven，执行以下命令下载依赖：

```bash
# 下载所有依赖
mvn dependency:resolve

# 或者直接编译（会自动下载依赖）
mvn clean compile

# 打包应用（包含下载依赖）
mvn clean package
```

### 下载依赖（Gradle）

如果使用 Gradle，执行以下命令下载依赖：

```bash
# 下载所有依赖
./gradlew dependencies --refresh-dependencies

# 或者直接构建（会自动下载依赖）
./gradlew build

# Windows 系统使用
gradlew.bat build
```

## 测试配置

创建测试类验证日志输出：

```java
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Application {
    private static final Logger logger = LoggerFactory.getLogger(Application.class);
    
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
        
        // 测试日志输出
        logger.info("Application started successfully");
        logger.warn("This is a warning message");
        logger.error("This is an error message");
        
        // 测试异常日志
        try {
            throw new RuntimeException("Test exception");
        } catch (Exception e) {
            logger.error("Exception occurred", e);
        }
    }
}
```

## 故障排查

### 1. 日志未发送到 MemoryLogMonitor

- 检查 MemoryLogMonitor 服务是否启动
- 检查端口 9090 是否被占用
- 检查防火墙设置
- 查看 Log4j2 状态日志（设置 `status="DEBUG"`）

### 2. 连接失败异常影响主程序

确保配置了 `<IgnoreExceptions>true</IgnoreExceptions>`，这样即使连接失败也不会影响主程序运行。

### 3. 日志格式问题

确保日志格式以换行符 `%n` 结尾，MemoryLogMonitor 按行接收日志。

### 4. 性能问题

如果日志量很大，建议：
- 使用 `AsyncAppender` 异步发送
- 调整日志级别，减少不必要的日志
- 使用过滤器过滤特定日志

## 完整示例

### application.yml

```yaml
logging:
  config: classpath:log4j2.xml
  level:
    root: INFO
    com.yourcompany: DEBUG
```

### log4j2.xml（完整配置）

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN" monitorInterval="30">
    <Properties>
        <Property name="LOG_PATTERN">%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n</Property>
        <Property name="LOG_HOST">${env:LOG_HOST:-localhost}</Property>
        <Property name="LOG_PORT">${env:LOG_PORT:-9090}</Property>
    </Properties>
    
    <Appenders>
        <!-- 控制台输出 -->
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="${LOG_PATTERN}"/>
        </Console>
        
        <!-- 文件输出（可选） -->
        <RollingFile name="FileAppender" 
                     fileName="logs/application.log"
                     filePattern="logs/application-%d{yyyy-MM-dd}-%i.log.gz">
            <PatternLayout pattern="${LOG_PATTERN}"/>
            <Policies>
                <TimeBasedTriggeringPolicy interval="1" modulate="true"/>
                <SizeBasedTriggeringPolicy size="100MB"/>
            </Policies>
            <DefaultRolloverStrategy max="10"/>
        </RollingFile>
        
        <!-- TCP Socket Appender - MemoryLogMonitor -->
        <Socket name="MemoryLogMonitorSocket" 
                host="${LOG_HOST}" 
                port="${LOG_PORT}" 
                protocol="TCP">
            <PatternLayout pattern="${LOG_PATTERN}"/>
            <IgnoreExceptions>true</IgnoreExceptions>
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <ImmediateFlush>true</ImmediateFlush>
            <!-- 连接超时（毫秒） -->
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </Socket>
        
        <!-- 异步包装器 -->
        <Async name="MemoryLogMonitor" bufferSize="1024">
            <AppenderRef ref="MemoryLogMonitorSocket"/>
            <IgnoreExceptions>true</IgnoreExceptions>
        </Async>
    </Appenders>
    
    <Loggers>
        <!-- 特定包的日志级别 -->
        <Logger name="com.yourcompany" level="DEBUG" additivity="false">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="FileAppender"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Logger>
        
        <!-- 根日志配置 -->
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="FileAppender"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

## 注意事项

1. **异常处理**: 务必设置 `<IgnoreExceptions>true</IgnoreExceptions>`，避免日志发送失败影响主程序
2. **网络延迟**: TCP 连接可能存在延迟，建议使用异步方式发送日志
3. **日志格式**: 确保日志格式以换行符结尾，MemoryLogMonitor 按行接收
4. **连接重试**: 配置合理的重试间隔，避免频繁重连
5. **日志级别**: 根据实际需求调整日志级别，避免产生过多日志

## 完整配置范例

### log4j2.xml 完整范例

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN" monitorInterval="30">
    <Properties>
        <Property name="LOG_PATTERN">%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n</Property>
        <Property name="LOG_HOST">${env:LOG_HOST:-localhost}</Property>
        <Property name="LOG_PORT">${env:LOG_PORT:-9090}</Property>
    </Properties>
    
    <Appenders>
        <!-- 控制台输出 -->
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="${LOG_PATTERN}"/>
        </Console>
        
        <!-- 文件输出（可选） -->
        <RollingFile name="FileAppender" 
                     fileName="logs/application.log"
                     filePattern="logs/application-%d{yyyy-MM-dd}-%i.log.gz">
            <PatternLayout pattern="${LOG_PATTERN}"/>
            <Policies>
                <TimeBasedTriggeringPolicy interval="1" modulate="true"/>
                <SizeBasedTriggeringPolicy size="100MB"/>
            </Policies>
            <DefaultRolloverStrategy max="10"/>
        </RollingFile>
        
        <!-- TCP Socket Appender - MemoryLogMonitor -->
        <Socket name="MemoryLogMonitorSocket" 
                host="${LOG_HOST}" 
                port="${LOG_PORT}" 
                protocol="TCP">
            <PatternLayout pattern="${LOG_PATTERN}"/>
            <IgnoreExceptions>true</IgnoreExceptions>
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <ImmediateFlush>true</ImmediateFlush>
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </Socket>
        
        <!-- 异步包装器（可选，提高性能） -->
        <Async name="MemoryLogMonitor" bufferSize="1024">
            <AppenderRef ref="MemoryLogMonitorSocket"/>
            <IgnoreExceptions>true</IgnoreExceptions>
        </Async>
    </Appenders>
    
    <Loggers>
        <!-- 特定包的日志级别 -->
        <Logger name="com.yourcompany" level="DEBUG" additivity="false">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="FileAppender"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Logger>
        
        <!-- 根日志配置 -->
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="FileAppender"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

### log4j2.properties 完整范例

```properties
# Log4j2 配置文件 - Properties 格式
# 适用于 Spring Boot 项目

# 状态日志级别
status = WARN
monitorInterval = 30

# 属性定义
property.LOG_PATTERN = %d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n
property.LOG_HOST = ${env:LOG_HOST:-localhost}
property.LOG_PORT = ${env:LOG_PORT:-9090}

# Appenders 配置
appenders = Console, FileAppender, MemoryLogMonitorSocket, MemoryLogMonitor

# 控制台输出
appender.Console.type = Console
appender.Console.name = Console
appender.Console.target = SYSTEM_OUT
appender.Console.layout.type = PatternLayout
appender.Console.layout.pattern = ${LOG_PATTERN}

# 文件输出（可选）
appender.FileAppender.type = RollingFile
appender.FileAppender.name = FileAppender
appender.FileAppender.fileName = logs/application.log
appender.FileAppender.filePattern = logs/application-%d{yyyy-MM-dd}-%i.log.gz
appender.FileAppender.layout.type = PatternLayout
appender.FileAppender.layout.pattern = ${LOG_PATTERN}
appender.FileAppender.policies.type = Policies
appender.FileAppender.policies.time.type = TimeBasedTriggeringPolicy
appender.FileAppender.policies.time.interval = 1
appender.FileAppender.policies.time.modulate = true
appender.FileAppender.policies.size.type = SizeBasedTriggeringPolicy
appender.FileAppender.policies.size.size = 100MB
appender.FileAppender.strategy.type = DefaultRolloverStrategy
appender.FileAppender.strategy.max = 10

# TCP Socket Appender - MemoryLogMonitor
appender.MemoryLogMonitorSocket.type = Socket
appender.MemoryLogMonitorSocket.name = MemoryLogMonitorSocket
appender.MemoryLogMonitorSocket.host = ${LOG_HOST}
appender.MemoryLogMonitorSocket.port = ${LOG_PORT}
appender.MemoryLogMonitorSocket.protocol = TCP
appender.MemoryLogMonitorSocket.layout.type = PatternLayout
appender.MemoryLogMonitorSocket.layout.pattern = ${LOG_PATTERN}
appender.MemoryLogMonitorSocket.ignoreExceptions = true
appender.MemoryLogMonitorSocket.reconnectDelayMillis = 30000
appender.MemoryLogMonitorSocket.immediateFlush = true
appender.MemoryLogMonitorSocket.connectTimeoutMillis = 1000

# 异步包装器（可选，提高性能）
appender.MemoryLogMonitor.type = Async
appender.MemoryLogMonitor.name = MemoryLogMonitor
appender.MemoryLogMonitor.bufferSize = 1024
appender.MemoryLogMonitor.appenderRef = MemoryLogMonitorSocket
appender.MemoryLogMonitor.ignoreExceptions = true

# Loggers 配置
loggers = Root, com.yourcompany

# 特定包的日志级别
logger.com.yourcompany.name = com.yourcompany
logger.com.yourcompany.level = DEBUG
logger.com.yourcompany.additivity = false
logger.com.yourcompany.appenderRef.Console.ref = Console
logger.com.yourcompany.appenderRef.FileAppender.ref = FileAppender
logger.com.yourcompany.appenderRef.MemoryLogMonitor.ref = MemoryLogMonitor

# 根日志配置
rootLogger.level = INFO
rootLogger.appenderRef.Console.ref = Console
rootLogger.appenderRef.FileAppender.ref = FileAppender
rootLogger.appenderRef.MemoryLogMonitor.ref = MemoryLogMonitor
```

### 简化版 log4j2.xml（仅控制台和 MemoryLogMonitor）

如果只需要控制台和 MemoryLogMonitor 输出，可以使用以下简化配置：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Configuration status="WARN">
    <Appenders>
        <!-- 控制台输出 -->
        <Console name="Console" target="SYSTEM_OUT">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
        </Console>
        
        <!-- TCP Socket Appender - MemoryLogMonitor -->
        <Socket name="MemoryLogMonitor" 
                host="localhost" 
                port="9090" 
                protocol="TCP">
            <PatternLayout pattern="%d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n"/>
            <IgnoreExceptions>true</IgnoreExceptions>
            <ReconnectDelayMillis>30000</ReconnectDelayMillis>
            <ImmediateFlush>true</ImmediateFlush>
            <ConnectTimeoutMillis>1000</ConnectTimeoutMillis>
        </Socket>
    </Appenders>
    
    <Loggers>
        <Root level="INFO">
            <AppenderRef ref="Console"/>
            <AppenderRef ref="MemoryLogMonitor"/>
        </Root>
    </Loggers>
</Configuration>
```

### 简化版 log4j2.properties（仅控制台和 MemoryLogMonitor）

```properties
# Log4j2 简化配置 - Properties 格式

status = WARN

# Appenders
appenders = Console, MemoryLogMonitor

# 控制台输出
appender.Console.type = Console
appender.Console.name = Console
appender.Console.target = SYSTEM_OUT
appender.Console.layout.type = PatternLayout
appender.Console.layout.pattern = %d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n

# TCP Socket Appender - MemoryLogMonitor
appender.MemoryLogMonitor.type = Socket
appender.MemoryLogMonitor.name = MemoryLogMonitor
appender.MemoryLogMonitor.host = localhost
appender.MemoryLogMonitor.port = 9090
appender.MemoryLogMonitor.protocol = TCP
appender.MemoryLogMonitor.layout.type = PatternLayout
appender.MemoryLogMonitor.layout.pattern = %d{yyyy-MM-dd HH:mm:ss.SSS} [%t] %-5level %logger{36} - %msg%n
appender.MemoryLogMonitor.ignoreExceptions = true
appender.MemoryLogMonitor.reconnectDelayMillis = 30000
appender.MemoryLogMonitor.immediateFlush = true
appender.MemoryLogMonitor.connectTimeoutMillis = 1000

# Root Logger
rootLogger.level = INFO
rootLogger.appenderRef.Console.ref = Console
rootLogger.appenderRef.MemoryLogMonitor.ref = MemoryLogMonitor
```

## 配置文件选择

- **log4j2.xml**: 更易读，支持更复杂的配置，推荐使用
- **log4j2.properties**: 更简洁，适合简单配置

两种格式功能相同，选择一种即可。Spring Boot 会按以下顺序查找配置文件：
1. `log4j2.xml`
2. `log4j2.properties`
3. `log4j2.json`
4. `log4j2.yaml`

## 参考资源

- [Log4j2 官方文档](https://logging.apache.org/log4j/2.x/manual/appenders.html#SocketAppender)
- [Spring Boot Logging](https://docs.spring.io/spring-boot/docs/current/reference/html/features.html#features.logging)
- [MemoryLogMonitor README](./README.md)
