"""Official Python SDKs against Tollgate: only the base URL changes (spec 002 FR-013, SC-001).

Environment: OPENAI_BASE_URL (ending in /v1) and ANTHROPIC_BASE_URL (without /v1).
The keys are fake: the simulators accept any value and never call a real provider.
"""

import os
import sys

import anthropic
import openai

FAKE_KEY = "tg-example-sdk-compat"


def check(cond: bool, msg: str) -> None:
    if not cond:
        raise AssertionError(msg)


def openai_suite(base_url: str) -> None:
    client = openai.OpenAI(base_url=base_url, api_key=FAKE_KEY, max_retries=0)
    msgs = [{"role": "user", "content": "Say hello."}]

    resp = client.chat.completions.create(model="gpt-4o-mini", messages=msgs, max_tokens=32)
    u = resp.usage
    check(bool(resp.choices[0].message.content), "chat: empty content")
    check(u is not None and u.total_tokens == u.prompt_tokens + u.completion_tokens, "chat: usage mismatch")

    text, usage = "", None
    stream = client.chat.completions.create(
        model="gpt-4o-mini", messages=msgs, stream=True, stream_options={"include_usage": True}
    )
    for chunk in stream:
        if chunk.choices and chunk.choices[0].delta.content:
            text += chunk.choices[0].delta.content
        if chunk.usage is not None:
            usage = chunk.usage
    check(bool(text) and usage is not None and usage.completion_tokens > 0, "chat stream: no text or usage")

    tool = {"type": "function", "function": {"name": "get_weather", "parameters": {"type": "object"}}}
    resp = client.chat.completions.create(model="gpt-4o-mini", messages=msgs, tools=[tool], tool_choice="required")
    calls = resp.choices[0].message.tool_calls or []
    check(len(calls) == 1 and calls[0].function.name == "get_weather", "chat: forced tool call missing")

    emb = client.embeddings.create(model="text-embedding-3-small", input=["first", "second"])
    check(len(emb.data) == 2 and len(emb.data[0].embedding) == 1536, "embeddings: wrong shape")

    ids = [m.id for m in client.models.list()]
    check("gpt-4o-mini" in ids, "models: gpt-4o-mini not listed")
    print("PASS python openai", openai.__version__)


def anthropic_suite(base_url: str) -> None:
    client = anthropic.Anthropic(base_url=base_url, api_key=FAKE_KEY, max_retries=0)
    msgs = [{"role": "user", "content": "Say hello."}]

    msg = client.messages.create(model="claude-sonnet-5-5", max_tokens=32, messages=msgs)
    check(msg.content[0].type == "text" and bool(msg.content[0].text), "messages: empty text")
    check(msg.usage.input_tokens > 0 and msg.usage.output_tokens > 0, "messages: usage missing")

    text, stop, output = "", None, 0
    for event in client.messages.create(model="claude-sonnet-5-5", max_tokens=32, messages=msgs, stream=True):
        if event.type == "content_block_delta" and event.delta.type == "text_delta":
            text += event.delta.text
        elif event.type == "message_delta":
            stop, output = event.delta.stop_reason, event.usage.output_tokens
    check(bool(text) and stop in ("end_turn", "max_tokens") and output > 0, "messages stream: incomplete")
    print("PASS python anthropic", anthropic.__version__)


def main() -> int:
    try:
        openai_suite(os.environ["OPENAI_BASE_URL"])
        anthropic_suite(os.environ["ANTHROPIC_BASE_URL"])
    except Exception as exc:  # report and fail the suite
        print(f"FAIL python: {exc!r}")
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
