# Feature Specification: Kubernetes Helm Deployment

## User Scenarios & Testing

### User Story 1 - Install gateway

An operator renders a pinned Helm chart with configurable image, replicas, service and resources.

### User Story 2 - Health-aware rollout

Kubernetes uses gateway health/readiness endpoints to route only ready pods.

## Functional Requirements

- FR-K8S-1: Chart MUST render a Deployment and Service without unpinned image tags.
- FR-K8S-2: Deployment MUST expose liveness and readiness probes.
- FR-K8S-3: Chart defaults MUST be safe for local rendering.

## Success Criteria

- SC-001: `helm lint` and `helm template` pass.
