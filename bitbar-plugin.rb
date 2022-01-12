#!/usr/bin/env ruby

require "json"
require "net/http"

TINY_IP = "192.168.164.241"
URL = "http://up-or-not.pickardayune.com:8080/api/target/8.8.8.8"

def main
  http = Net::HTTP.new(TINY_IP, 8080)
  http.open_timeout = 1.0
  http.start do |http|
    req = Net::HTTP::Get.new(URI(URL))
    res = http.request(req)
    case res
    when Net::HTTPOK
      data = JSON.parse(res.body)
      printf "#{score(data)}\n---\n#{summarize(data)}\n"
    else
      printf "❓\n---\n#{res.code}\n"
      return
    end
  end
rescue => e
  printf "❓\n---\n#{e}\n"
end

RED = "🔴"
YELLOW = "🟡"
GREEN = "🟢"

def status(color, msg)
  "#{color} #{msg}"
end

def score(data)
  # If there are no ping responses, the connection is not working.
  if data["ok"] == 0
    return status(RED, "offline")
  end

  # If some of the ping responses are missing, the connection is not good.
  missing = data["count"] - data["ok"]
  if missing > 0
    ratio = data["count"] / missing
    if ratio < 2 # at least 25/50 missing
      return status(RED, "lossy")
    elsif ratio < 15 # at least 3/50 missing
      return status(YELLOW, "lossy")
    end
  end

  # If the average ping time is over 100ms, we'll notice it being slow.
  if data["avg_ms"] > 100.0
    return status(YELLOW, "laggy %.0fms" % data["avg_ms"])
  end

  # See if at least 2/3 of the packets come back in the fastest bucket (60ms).
  # Alternatively, allow half of the next slowest bucket to count towards this
  # limit. In other words, if 25% are fast and 50% are slow, things are
  # probably OK.
  fast = data["buckets"][0]["count"]
  not_too_slow = data["buckets"][1]["count"]
  weighted = fast + (not_too_slow / 2)
  limit = 2 * data["count"] / 3
  if weighted < limit
    return status(YELLOW, "slow (#{weighted}/#{limit})")
  end

  # If none of the above are true, then things must be good!
  GREEN
end

def summarize(data)
  lines = [
    "ok: #{data["ok"]}/#{data["count"]}",
    "avg: #{data["avg_ms"]} ms",
  ]
  data["buckets"].each do |bucket|
    label = bucket["max_ms"] ? "#{bucket["max_ms"]} ms" : "slower"
    lines << "#{label}: #{bucket["count"]}"
  end
  lines.join("\n")
end

main
