# Feature Specification: Response Cache

## User Scenarios & Testing

### User Story 1 - Exact safe caching

The gateway reuses a response for an identical method, path and request body while respecting a bounded TTL.

### User Story 2 - Isolation

Requests with different bodies never share cached responses, and expired entries are ignored.

## Functional Requirements

- FR-CACHE-1: Keys MUST include method, path and full request body.
- FR-CACHE-2: Entries MUST have bounded TTL and copied response bodies.
- FR-CACHE-3: Cache metadata MUST exclude prompt and response content from logs.

## Success Criteria

- SC-001: Identical requests hit; different requests miss.
- SC-002: Expired entries never return.
