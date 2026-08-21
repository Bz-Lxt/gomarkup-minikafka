#!/usr/bin/env python3
"""API smoke tests for MiniKafka (Mock/offline, Cost ¥0)."""
from __future__ import annotations

import json
import os
import sys
import time
import urllib.error
import urllib.request

BASE = os.environ.get("BASE_URL", "http://127.0.0.1:18591").rstrip("/")
failures = []
TOPIC = "smoke-orders-" + str(int(time.time()))
GROUP = "smoke-group-" + str(int(time.time()))


def req(method: str, path: str, body=None, expect=200):
    data = None
    headers = {}
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    r = urllib.request.Request(BASE + path, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(r, timeout=10) as resp:
            raw = resp.read().decode()
            code = resp.status
            parsed = json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        raw = e.read().decode()
        code = e.code
        try:
            parsed = json.loads(raw)
        except json.JSONDecodeError:
            parsed = {"raw": raw}
    if code != expect:
        failures.append(f"{method} {path} -> {code} want {expect}: {parsed}")
        return parsed
    return parsed


def main():
    h = req("GET", "/health", expect=200)
    if h.get("status") != "ok":
        failures.append(f"health {h}")

    page = urllib.request.urlopen(BASE + "/", timeout=10)
    html = page.read().decode()
    if "MiniKafka" not in html and "观测台" not in html and "app" not in html:
        failures.append("SPA index missing MiniKafka markers")

    t = req("POST", "/api/v1/topics", {"name": TOPIC, "partitions": 2}, expect=201)
    if t.get("data", {}).get("name") != TOPIC:
        failures.append(f"create topic {t}")

    topics = req("GET", "/api/v1/topics")
    names = [x["name"] for x in topics.get("data", [])]
    if TOPIC not in names:
        failures.append(f"topic list {topics}")

    produced = req(
        "POST",
        "/api/v1/produce/batch",
        {"topic": TOPIC, "messages": [{"key": "a", "value": "1"}, {"key": "b", "value": "2"}]},
        expect=201,
    )
    if produced.get("data", {}).get("count", 0) < 1:
        failures.append(f"produce {produced}")

    consumed = req(
        "POST",
        "/api/v1/consume",
        {
            "topic": TOPIC,
            "group": GROUP,
            "client_id": "smoke-c1",
            "max_messages": 10,
            "auto_commit": True,
        },
        expect=200,
    )
    msgs = consumed.get("data", {}).get("messages", [])
    if not msgs:
        failures.append(f"consume empty {consumed}")

    groups = req("GET", "/api/v1/groups")
    gnames = [g["group"] for g in groups.get("data", [])]
    if GROUP not in gnames:
        failures.append(f"groups {groups}")

    metrics = req("GET", "/api/v1/metrics")
    if "messages_total" not in metrics.get("data", {}):
        failures.append(f"metrics {metrics}")

    msgs2 = req("GET", f"/api/v1/topics/{TOPIC}/messages?partition=0&offset=0&limit=10")
    if "messages" not in msgs2.get("data", {}):
        failures.append(f"browse {msgs2}")

    req(
        "POST",
        f"/api/v1/groups/{GROUP}/reset",
        {"topic": TOPIC, "to": "earliest"},
        expect=200,
    )

    if failures:
        print("SMOKE FAIL")
        for f in failures:
            print(" -", f)
        sys.exit(1)
    print("SMOKE PASS")
    print("[PASS] Health Check")
    print("[PASS] SPA")
    print("[PASS] Topic / Produce / Consume / Offset / Metrics")


if __name__ == "__main__":
    main()
