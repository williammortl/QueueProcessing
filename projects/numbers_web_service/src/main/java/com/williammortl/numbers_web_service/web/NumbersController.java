package com.williammortl.numbers_web_service.web;

import com.williammortl.numbers_web_service.service.NumbersStore;
import io.swagger.v3.oas.annotations.Operation;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.Instant;
import java.util.List;
import java.util.Map;

@RestController
public class NumbersController {
    private final NumbersStore store;

    public NumbersController(NumbersStore store) { this.store = store; }

    @Operation(summary = "Heartbeat")
    @GetMapping("/ping")
    public Map<String, Object> ping() {
        return Map.of("msg", "pong!", "time", Instant.now().toString());
    }

    @Operation(summary = "Get recent numbers")
    @GetMapping("/numbers")
    public Map<String, Object> numbers() {
        return Map.of("numbers", store.latest(100));
    }

    @Operation(summary = "Get average of most recent numbers (up to 100)")
    @GetMapping("/average")
    public Map<String, Object> average() {
        List<Long> recent = store.latest(100);
        double avg = 0.0;
        if (!recent.isEmpty()) {
            long sum = 0L;
            for (Long n : recent) sum += n;
            avg = (double) sum / recent.size();
        }
        return Map.of("average", avg);
    }

    @Operation(summary = "Get Nth stored number (0-based index)")
    @GetMapping("/number/{n}")
    public ResponseEntity<Map<String, Object>> number(@PathVariable("n") int n) {
        try {
            long value = store.getByIndex(n);
            return ResponseEntity.ok(Map.of("index", n, "number", value));
        } catch (IndexOutOfBoundsException ex) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body(Map.of("error", "index out of range"));
        }
    }

    public record NumbersPayload(List<Long> numbers) {}

    @Operation(summary = "Replace oldest M numbers with given list")
    @PutMapping("/numbers")
    public Map<String, Object> putNumbers(@RequestBody NumbersPayload payload) {
        store.putAll(payload.numbers());
        return Map.of("stored", store.size());
    }
}
