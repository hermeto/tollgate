# Implementation Plan: Routing and Fallback

Provide a local weighted selector and bounded backoff abstraction. Transport integration can use the
same selector for provider targets and circuit state can be layered without changing callers.
