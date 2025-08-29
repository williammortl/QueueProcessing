# Numbers Web Service

Spring Boot (Java 24) + Gradle 9.0.0. Exposes Swagger and Prometheus metrics.

## Build

```
gradle wrapper --gradle-version 9.0
./gradlew clean build
```

Executable JAR: build/libs/numbers-web-service-0.1.0.jar

## Run

```
java -jar build/libs/numbers-web-service-0.1.0.jar --port=8080 --metrics=9090 --n=100
```

- Swagger UI: /swagger-ui.html
- Metrics: /actuator/prometheus
