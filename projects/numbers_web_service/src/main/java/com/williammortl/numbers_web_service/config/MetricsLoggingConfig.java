package com.williammortl.numbers_web_service.config;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.MeterRegistry;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.lang.NonNull;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.time.Instant;

@Configuration
public class MetricsLoggingConfig {
    private static final Logger log = LoggerFactory.getLogger(MetricsLoggingConfig.class);
    private final MeterRegistry registry;
    private Counter requests;

    public MetricsLoggingConfig(MeterRegistry registry) {
        this.registry = registry;
    }

    @PostConstruct
    public void start() {
        registry.gauge("service_start_time_seconds", Instant.now().getEpochSecond());
        requests = Counter.builder("http_requests_total").description("Total HTTP requests").register(registry);
        log.info("web service startup");
    }

    @PreDestroy
    public void stop() {
        registry.gauge("service_shutdown_time_seconds", Instant.now().getEpochSecond());
        log.info("web service shutdown");
    }

    @Bean
    public OncePerRequestFilter requestLogFilter() {
        return new OncePerRequestFilter() {
            @Override
            protected void doFilterInternal(@NonNull HttpServletRequest request, @NonNull HttpServletResponse response, @NonNull FilterChain filterChain) throws ServletException, IOException {
                String method = request.getMethod();
                String path = request.getRequestURI();
                log.info("{} {}", method, path);
                requests.increment();
                filterChain.doFilter(request, response);
            }
        };
    }
}
