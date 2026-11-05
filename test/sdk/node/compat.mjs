// Official TypeScript SDKs against Tollgate: only the base URL changes (spec 002 FR-013, SC-001).
// Environment: OPENAI_BASE_URL (ending in /v1) and ANTHROPIC_BASE_URL (without /v1). Keys are fake.
import Anthropic from "@anthropic-ai/sdk";
import OpenAI from "openai";

const FAKE_KEY = "tg-example-sdk-compat";

function check(cond, msg) {
  if (!cond) throw new Error(msg);
}

async function openaiSuite(baseURL) {
  const client = new OpenAI({ baseURL, apiKey: FAKE_KEY, maxRetries: 0 });
  const messages = [{ role: "user", content: "Say hello." }];

  const resp = await client.chat.completions.create({ model: "gpt-4o-mini", messages, max_tokens: 32 });
  const u = resp.usage;
  check(resp.choices[0].message.content, "chat: empty content");
  check(u && u.total_tokens === u.prompt_tokens + u.completion_tokens, "chat: usage mismatch");

  let text = "";
  let usage = null;
  const stream = await client.chat.completions.create({
    model: "gpt-4o-mini", messages, stream: true, stream_options: { include_usage: true },
  });
  for await (const chunk of stream) {
    text += chunk.choices[0]?.delta?.content ?? "";
    if (chunk.usage) usage = chunk.usage;
  }
  check(text && usage && usage.completion_tokens > 0, "chat stream: no text or usage");

  const tool = { type: "function", function: { name: "get_weather", parameters: { type: "object" } } };
  const forced = await client.chat.completions.create({ model: "gpt-4o-mini", messages, tools: [tool], tool_choice: "required" });
  const calls = forced.choices[0].message.tool_calls ?? [];
  check(calls.length === 1 && calls[0].function.name === "get_weather", "chat: forced tool call missing");

  const emb = await client.embeddings.create({ model: "text-embedding-3-small", input: ["first", "second"] });
  check(emb.data.length === 2 && emb.data[0].embedding.length === 1536, "embeddings: wrong shape");

  const ids = [];
  for await (const m of client.models.list()) ids.push(m.id);
  check(ids.includes("gpt-4o-mini"), "models: gpt-4o-mini not listed");
  console.log("PASS node openai", OpenAI.VERSION ?? "");
}

async function anthropicSuite(baseURL) {
  const client = new Anthropic({ baseURL, apiKey: FAKE_KEY, maxRetries: 0 });
  const messages = [{ role: "user", content: "Say hello." }];

  const msg = await client.messages.create({ model: "claude-sonnet-5-5", max_tokens: 32, messages });
  check(msg.content[0].type === "text" && msg.content[0].text, "messages: empty text");
  check(msg.usage.input_tokens > 0 && msg.usage.output_tokens > 0, "messages: usage missing");

  let text = "";
  let stop = null;
  let output = 0;
  const stream = await client.messages.create({ model: "claude-sonnet-5-5", max_tokens: 32, messages, stream: true });
  for await (const event of stream) {
    if (event.type === "content_block_delta" && event.delta.type === "text_delta") text += event.delta.text;
    if (event.type === "message_delta") {
      stop = event.delta.stop_reason;
      output = event.usage.output_tokens;
    }
  }
  check(text && ["end_turn", "max_tokens"].includes(stop) && output > 0, "messages stream: incomplete");
  console.log("PASS node anthropic", Anthropic.VERSION ?? "");
}

try {
  await openaiSuite(process.env.OPENAI_BASE_URL);
  await anthropicSuite(process.env.ANTHROPIC_BASE_URL);
} catch (err) {
  console.log(`FAIL node: ${err}`);
  process.exit(1);
}
