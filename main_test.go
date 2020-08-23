package main

import (
    "testing"
)

func TestGetIntervalTime_10(t *testing.T) {
    interval_time := getIntervalTime(10)
    if interval_time != 1 {
        t.Error("invalid interval")
    }
}

func TestGetIntervalTime_30(t *testing.T) {
    interval_time := getIntervalTime(30)
    if interval_time != 24 {
        t.Error("invalid interval")
    }
}

func TestGetIntervalTime_50(t *testing.T) {
    interval_time := getIntervalTime(50)
    if interval_time != 48 {
        t.Error("invalid interval")
    }
}
