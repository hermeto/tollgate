# /// script
# requires-python = ">=3.11"
# dependencies = ["tiktoken==0.14.0"]
# ///
"""Writes reference token counts for testdata/tokens/corpus.jsonl using tiktoken o200k_base.

Run: uv run --locked --script scripts/tokens/reference.py
The output is committed as testdata/tokens/reference_o200k.json; a Go test compares
internal/tokens against it (spec 002 SC-006). The BPE file is downloaded once from
openaipublic.blob.core.windows.net (public, free) and cached by tiktoken.
"""

import json
import pathlib

import tiktoken

ROOT = pathlib.Path(__file__).resolve().parents[2]
CORPUS = ROOT / "testdata" / "tokens" / "corpus.jsonl"
OUT = ROOT / "testdata" / "tokens" / "reference_o200k.json"


def main() -> None:
    enc = tiktoken.get_encoding("o200k_base")
    rows = []
    for line in CORPUS.read_text(encoding="utf-8").splitlines():
        sample = json.loads(line)
        rows.append({"id": sample["id"], "tokens": len(enc.encode(sample["text"]))})
    doc = {"tokenizer": "o200k_base", "tiktoken": tiktoken.__version__, "samples": rows}
    OUT.write_text(json.dumps(doc, indent=2) + "\n", encoding="utf-8")
    print(f"wrote {len(rows)} reference counts to {OUT.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
