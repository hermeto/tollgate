// Time-to-first-byte accuracy of the simulator under 1,000 concurrent streams (spec 002 SC-007).
// Start the simulator with a fixed TTFT, for example:
//   tollgate sim --flavor openai --port 19202 --config bench/k6/sim-ttft.yaml &
//   k6 run -e BASE_URL=http://127.0.0.1:19202 bench/k6/sim-ttft.js
// VUs ramp up over 15 s (macOS listen backlogs reject bursts of new connections), then hold STREAMS
// (default 1,000) concurrent streams for 30 s; only the hold phase is reported
// (http_req_waiting{phase:hold}).
import http from "k6/http";
import exec from "k6/execution";

export const options = {
  scenarios: {
    streams: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "15s", target: Number(__ENV.STREAMS || 1000) },
        { duration: "30s", target: Number(__ENV.STREAMS || 1000) },
      ],
      gracefulRampDown: "5s",
    },
  },
  thresholds: {
    "http_req_waiting{phase:hold}": ["p(99)<105"],
    "http_req_failed{phase:hold}": ["rate<0.001"],
  },
  summaryTrendStats: ["min", "med", "p(90)", "p(99)", "max"],
};

const url = `${__ENV.BASE_URL || "http://127.0.0.1:9001"}/v1/chat/completions`;
const body = JSON.stringify({
  model: "gpt-4o-mini",
  stream: true,
  messages: [{ role: "user", content: "Say hello." }],
});

export default function () {
  const phase = exec.instance.currentTestRunDuration > 17000 ? "hold" : "ramp";
  http.post(url, body, {
    headers: { "Content-Type": "application/json", Authorization: "Bearer tg-example-k6" },
    tags: { phase },
  });
}
