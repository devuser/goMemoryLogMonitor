package com.memorylogmonitor.mock.controller;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.servlet.mvc.support.RedirectAttributes;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

@Controller
public class LogController {
    private static final Logger logger = LoggerFactory.getLogger(LogController.class);
    private static final DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSS");

    @GetMapping("/")
    public String index() {
        return "index";
    }

    @PostMapping("/send-log")
    public String sendLog(
            @RequestParam("logMessage") String logMessage,
            @RequestParam(value = "logLevel", defaultValue = "INFO") String logLevel,
            RedirectAttributes redirectAttributes) {
        
        if (logMessage == null || logMessage.trim().isEmpty()) {
            redirectAttributes.addFlashAttribute("error", "日志内容不能为空");
            return "redirect:/";
        }

        try {
            // 格式化日志消息
            String timestamp = LocalDateTime.now().format(formatter);
            String formattedLog = String.format("[%s] [%s] %s", timestamp, logLevel, logMessage);
            
            // 根据日志级别发送日志
            switch (logLevel.toUpperCase()) {
                case "DEBUG":
                    logger.debug(formattedLog);
                    break;
                case "WARN":
                    logger.warn(formattedLog);
                    break;
                case "ERROR":
                    logger.error(formattedLog);
                    break;
                case "INFO":
                default:
                    logger.info(formattedLog);
                    break;
            }
            
            redirectAttributes.addFlashAttribute("success", "日志已成功发送到 MemoryLogMonitor");
        } catch (Exception e) {
            logger.error("发送日志失败", e);
            redirectAttributes.addFlashAttribute("error", "发送日志失败: " + e.getMessage());
        }

        return "redirect:/";
    }
}
