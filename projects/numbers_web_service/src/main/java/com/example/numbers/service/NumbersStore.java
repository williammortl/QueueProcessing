package com.example.numbers.service;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.atomic.AtomicLong;

@Service
public class NumbersStore {
    private final int capacity;
    private final long[] buffer;
    private int writeIdx = 0;
    private int size = 0;
    private final AtomicLong count = new AtomicLong(0);
    private final AtomicLong sum = new AtomicLong(0);

    public NumbersStore(@Value("${numbers.n}") int capacity) {
        this.capacity = capacity;
        this.buffer = new long[capacity];
    }

    public synchronized void putAll(List<Long> numbers) {
        for (Long n : numbers) put(n);
    }

    public synchronized void put(long n) {
        if (size < capacity) {
            buffer[writeIdx] = n;
            size++;
        } else {
            buffer[writeIdx] = n;
        }
        writeIdx = (writeIdx + 1) % capacity;
        count.incrementAndGet();
        sum.addAndGet(n);
    }

    public synchronized List<Long> latest(int k) {
        int toReturn = Math.min(k, size);
        List<Long> out = new ArrayList<>(toReturn);
        for (int i = 0; i < toReturn; i++) {
            int idx = (writeIdx - 1 - i + capacity) % capacity;
            out.add(buffer[idx]);
        }
        Collections.reverse(out);
        return out;
    }

    public synchronized long getByIndex(int idx) {
        if (idx < 0 || idx >= size) throw new IndexOutOfBoundsException();
        int start = (writeIdx - size + capacity) % capacity;
        int real = (start + idx) % capacity;
        return buffer[real];
    }

    public double averageAllTime() {
        long c = count.get();
        return c == 0 ? 0.0 : ((double) sum.get()) / c;
    }

    public synchronized int size() { return size; }
    public int capacity() { return capacity; }
}
