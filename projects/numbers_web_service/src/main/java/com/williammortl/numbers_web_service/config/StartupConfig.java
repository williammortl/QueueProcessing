package com.williammortl.numbers_web_service.config;

import jakarta.annotation.PostConstruct;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.env.ConfigurableEnvironment;

@Configuration
@ConfigurationProperties(prefix = "numbers")
public class StartupConfig {
    private static final Logger log = LoggerFactory.getLogger(StartupConfig.class);
    private final ApplicationArguments args;
    private final ConfigurableEnvironment env;

    @Value("${numbers.n:0}")
    private int n;

    public StartupConfig(ApplicationArguments args, ConfigurableEnvironment env) {
        this.args = args;
        this.env = env;
    }

    @PostConstruct
    public void init() {
        Integer httpPort = getIntArg("port");
        Integer metricsPort = getIntArg("metrics");
        Integer cap = getIntArg("n");

        if (httpPort == null || metricsPort == null || cap == null || cap <= 0) {
            System.err.println("\nUsage: java -jar app.jar --port=<httpPort> --metrics=<prometheusPort> --n=<capacity>\n");
            throw new IllegalArgumentException("Missing required parameters");
        }

        env.getSystemProperties().put("server.port", String.valueOf(httpPort));
        env.getSystemProperties().put("management.server.port", String.valueOf(metricsPort));
        env.getSystemProperties().put("numbers.n", String.valueOf(cap));

        log.info("startup parameters: server.port={} management.server.port={} n={} ", httpPort, metricsPort, cap);
    }

    private Integer getIntArg(String name) {
        if (args.containsOption(name)) {
            try {
                return Integer.parseInt(args.getOptionValues(name).get(0));
            } catch (Exception ignored) { }
        }
        return null;
    }
}
