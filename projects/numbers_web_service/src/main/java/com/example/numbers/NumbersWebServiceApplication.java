package com.example.numbers;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.core.env.PropertiesPropertySource;
import org.springframework.core.env.StandardEnvironment;

import java.util.Properties;

@SpringBootApplication
public class NumbersWebServiceApplication {
    public static void main(String[] args) {
        // Parse minimal args before boot
        Integer http = getIntArg(args, "--port=");
        Integer metrics = getIntArg(args, "--metrics=");
        Integer cap = getIntArg(args, "--n=");
        if (http == null || metrics == null || cap == null || cap <= 0) {
            System.err.println("\nUsage: java -jar app.jar --port=<httpPort> --metrics=<prometheusPort> --n=<capacity>\n");
            System.err.println("\nby William Mortl\n");
            System.exit(2);
        }

        Properties props = new Properties();
        props.setProperty("server.port", String.valueOf(http));
        props.setProperty("management.server.port", String.valueOf(metrics));
        props.setProperty("numbers.n", String.valueOf(cap));

        SpringApplication app = new SpringApplication(NumbersWebServiceApplication.class);
        app.addInitializers(context -> {
            StandardEnvironment env = (StandardEnvironment) context.getEnvironment();
            env.getPropertySources().addFirst(new PropertiesPropertySource("cli", props));
        });
        app.run(args);
    }

    private static Integer getIntArg(String[] args, String prefix) {
        for (String a : args) {
            if (a.startsWith(prefix)) {
                try { return Integer.parseInt(a.substring(prefix.length())); } catch (Exception ignored) {}
            }
        }
        return null;
    }
}
