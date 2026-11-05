// Cross-check of simulator throughput (spec 002 SC-003) with an open-loop arrival rate.
// Usage: tollgate sim --flavor openai --port 19201 &
//        k6 run -e BASE_URL=http://127.0.0.1:19201 -e RATE=10000 -e DURATION=20s bench/k6/sim-throughput.js
import http from "k6/http";

export const options = {
  scenarios: {
    open_loop: {
      executor: "constant-arrival-rate",
      rate: Number(__ENV.RATE || 10000),
      timeUnit: "1s",
      duration: __ENV.DURATION || "20s",
      preAllocatedVUs: 200,
      maxVUs: 2000,
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.001"],
    dropped_iterations: ["count<100"],
  },
};

const url = `${__ENV.BASE_URL || "http://127.0.0.1:9001"}/v1/chat/completions`;
const body = JSON.stringify({
  model: "gpt-4o-mini",
  messages: [{ role: "user", content: "Say hello." }],
  max_tokens: 20,
});
const params = { headers: { "Content-Type": "application/json", Authorization: "Bearer tg-example-k6" } };

export default function () {
  http.post(url, body, params);
}
