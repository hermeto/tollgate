# Implementation Plan: Budgets and Chargeback

Use an in-memory reservation ledger for the simulator profile and a writer that maps content-free
usage events to stable FOCUS-aligned CSV columns. Redis/database adapters remain replaceable.
