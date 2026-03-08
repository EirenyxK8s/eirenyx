# Eirenyx

Eirenyx is a Kubernetes operator that acts as a unified control plane for security scanning, runtime threat detection, and resilience testing. It orchestrates existing Kubernetes-native tools — **Trivy**, **Falco**, and **Litmus** — through a single declarative interface.

## Overview

Modern Kubernetes clusters rely on multiple specialized tools for vulnerability scanning, runtime security, and resilience testing. These tools are typically deployed and operated independently, which leads to fragmented configuration, uncorrelated results, and limited automation.

Eirenyx introduces a unifying Kubernetes operator that manages the lifecycle of security and chaos tools and exposes a single policy-driven interface for triggering scans, detecting runtime threats, and executing resilience experiments.

Rather than replacing existing tools, Eirenyx coordinates them and aggregates their outputs into unified policy-level reports.

---

## Architecture

Eirenyx is built as a [Kubebuilder](https://book.kubebuilder.io/)-based Kubernetes operator. It defines three Custom Resource Definitions (CRDs) and runs three reconciliation controllers that coordinate the full lifecycle from tool installation to result reporting.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                       │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Eirenyx Operator                      │   │
│  │                                                          │   │
│  │  ┌────────────────┐  ┌────────────────┐  ┌───────────┐   │   │
│  │  │  Tool          │  │  Policy        │  │  Policy   │   │   │
│  │  │  Controller    │  │  Controller    │  │  Report   │   │   │
│  │  │                │  │                │  │  Control. │   │   │
│  │  └───────┬────────┘  └───────┬────────┘  └─────┬─────┘   │   │
│  │          │                   │                  │        │   │
│  │    ToolService          PolicyEngine       ReportHandler │   │
│  │    (per tool)           (per type)         (per type)    │   │
│  └──────────┼───────────────────┼──────────────────┼────────┘   │
│             │                   │                  │            │
│    ┌────────▼──────┐   ┌────────▼───────┐  ┌──────▼──────┐      │
│    │  Helm Manager │   │  K8s Resources │  │  Tool CRs   │      │
│    │  (install /   │   │  (Jobs, CMs,   │  │  (Trivy     │      │
│    │  uninstall)   │   │  ChaosEngines) │  │  VulnReport)│      │
│    └───────────────┘   └────────────────┘  └─────────────┘      │ 
│                                                                 │
│  ┌────────────┐  ┌─────────────────┐  ┌───────────────────┐     │
│  │  Trivy     │  │  Falco          │  │  Litmus           │     │
│  │  Operator  │  │  (DaemonSet)    │  │  ChaosCenter      │     │
│  │  (trivy-   │  │  (falco ns)     │  │  (litmus ns)      │     │
│  │   system)  │  │                 │  │                   │     │
│  └────────────┘  └─────────────────┘  └───────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

### Custom Resource Definitions

| CRD | API Group | Purpose |
|-----|-----------|---------|
| `Tool` | `eirenyx.eirenyx/v1alpha1` | Declares a security tool to install and manage via Helm |
| `Policy` | `eirenyx.eirenyx/v1alpha1` | Defines how an enabled tool should scan, monitor, or test workloads |
| `PolicyReport` | `eirenyx.eirenyx/v1alpha1` | Aggregates results produced by the tools for a given policy |

---

### CRD: Tool

A `Tool` resource represents the declarative installation and configuration of an external tool. Eirenyx uses Helm to install and manage the tool's lifecycle within the cluster.

**Supported tool types:** `trivy` | `falco` | `litmus`

```yaml
apiVersion: eirenyx.eirenyx/v1alpha1
kind: Tool
metadata:
  name: trivy
spec:
  type: trivy
  enabled: true
  namespace: trivy-system   # optional, uses default if omitted
  values: {}                # optional Helm values (raw JSON/YAML)
```

**Status fields:**

| Field | Description |
|-------|-------------|
| `installed` | Whether the Helm release is present |
| `healthy` | Whether the tool's workload (Deployment/DaemonSet) is ready |
| `version` | Installed chart version |
| `conditions` | Standard Kubernetes conditions (e.g. `Ready`) |

**Tool defaults:**

| Tool | Default Namespace | Workload Type | Health Check Target |
|------|-------------------|---------------|---------------------|
| Trivy Operator | `trivy-system` | Deployment | `trivy-operator` |
| Falco | `falco` | DaemonSet | `falco` |
| Litmus ChaosCenter | `litmus` | Deployment | `chaos-litmus-server` |

---

### CRD: Policy

A `Policy` resource defines what a tool should do against specific workloads. Its `type` field determines which tool engine is invoked. Each policy is automatically linked to its corresponding `Tool` as an owner reference, so deleting a `Tool` cascades to its policies.

**Supported policy types:** `trivy` | `falco` | `litmus`

#### Trivy Policy

Triggers one or more container image vulnerability scans by spawning Kubernetes `Job` resources. Each scan targets a specific image with configurable severity filters.

```yaml
apiVersion: eirenyx.eirenyx/v1alpha1
kind: Policy
metadata:
  name: scan-my-app
spec:
  type: trivy
  enabled: true
  trivy:
    scans:
      - name: app-scan
        image: my-registry/my-app:latest
        severity: CRITICAL,HIGH
        ignoreUnfixed: true
        vulnerabilityTypes:
          - os
          - library
```

Each scan produces a `Job` named `eirenyx-trivy-<policy>-<scan>` with a 30-minute TTL after completion.

#### Falco Policy

Configures Falco runtime threat detection by creating a `ConfigMap` that describes which Falco rules to observe. Rules can be referenced by name (`ruleRef`) or selected by tags/priority (`ruleSelector`).

```yaml
apiVersion: eirenyx.eirenyx/v1alpha1
kind: Policy
metadata:
  name: detect-shell-in-container
spec:
  type: falco
  enabled: true
  falco:
    observe:
      ruleRef:
        name: terminal-shell-in-container
    report:
      create: true
      severity: WARNING
      aggregationWindow: 5m
```

#### Litmus Policy

Runs chaos experiments against target workloads by creating `ChaosEngine` resources. Each experiment is configured with a target application, chaos type, duration, and expected outcome.

```yaml
apiVersion: eirenyx.eirenyx/v1alpha1
kind: Policy
metadata:
  name: pod-delete-experiment
spec:
  type: litmus
  enabled: true
  litmus:
    experiments:
      - name: pod-kill
        experimentRef: pod-delete
        appInfo:
          appNamespace: default
          appLabel: app=my-service
          appKind: deployment
        duration: "30"
        mode: Sequential
        parameters:
          FORCE: "false"
```

---

### CRD: PolicyReport

A `PolicyReport` is automatically created and managed by the `Policy` controller. Users do not create these directly. When a policy changes generation, the report is reset to `Pending` and re-evaluated.

**Report lifecycle phases:** `Pending` → `Running` → `Completed` / `Failed`

**Verdict values:** `Pass` | `Fail` | `Unknown`

```yaml
apiVersion: eirenyx.eirenyx/v1alpha1
kind: PolicyReport
metadata:
  name: trivy-report-scan-my-app
status:
  phase: Completed
  summary:
    verdict: Fail
    totalChecks: 42
    passed: 38
    failed: 4
  details: {}   # Raw JSON, structure varies by tool type
```

**Report details by tool type:**

| Tool | Details Structure |
|------|-------------------|
| Trivy | `{ image, vulnerabilities[], reportCount }` |
| Falco | `{ message, rule, podDetails[] }` |
| Litmus | Chaos experiment results from `ChaosResult` resources |

---

## Internal Architecture

### Package Layout

```
eirenyx/
├── cmd/
│   └── main.go                  # Operator entrypoint; registers controllers and schemes
├── api/
│   ├── v1alpha1/                # CRD type definitions (Tool, Policy, PolicyReport)
│   └── litmus/                  # Litmus CRD types (ChaosEngine, ChaosExperiment, ChaosResult)
└── internal/
    ├── controller/
    │   ├── tool_controller.go       # Reconciles Tool resources
    │   ├── policy_controller.go     # Reconciles Policy resources
    │   ├── policyreport_controller.go # Reconciles PolicyReport resources
    │   ├── factory.go               # Instantiates tool services, policy engines, report handlers
    │   └── rbac.go                  # RBAC marker annotations
    ├── tools/
    │   ├── tool.go                  # ToolService interface
    │   ├── trivy.go                 # Trivy Helm lifecycle service
    │   ├── falco.go                 # Falco Helm lifecycle service
    │   └── litmus.go                # Litmus Helm lifecycle service
    ├── policy/
    │   ├── engine.go                # PolicyEngine interface
    │   ├── trivy/engine.go          # Creates scan Jobs per TrivyScan spec
    │   ├── falco/engine.go          # Creates ConfigMaps for Falco rule observation
    │   └── litmus/engine.go         # Creates ChaosEngine resources per experiment
    ├── report/
    │   ├── handler.go               # ReportHandler interface
    │   ├── trivy.go                 # Reads trivy-operator VulnerabilityReports
    │   ├── falco.go                 # Aggregates Falco events from cluster
    │   ├── litmus.go                # Reads LitmusChaos ChaosResult resources
    │   └── pod_details.go           # Fetches pod metadata for report enrichment
    └── client/
        ├── helm/manager.go          # Helm SDK wrapper (install, upgrade, uninstall)
        └── k8s/client.go            # Kubernetes client helpers (namespace, deployment/daemonset health)
```

### Controller Reconciliation Flows

#### Tool Controller

```
Tool CR created/updated
  └─> Add finalizer
  └─> If enabled:  ToolService.EnsureInstalled  (Helm install/upgrade)
      If disabled: ToolService.EnsureUninstalled (Helm uninstall)
  └─> ToolService.CheckHealth (Deployment/DaemonSet readiness)
  └─> Update Tool status (installed, healthy)
  └─> If not healthy: requeue after 5s

Tool CR deleted
  └─> ToolService.EnsureUninstalled
  └─> Remove finalizer
```

#### Policy Controller

```
Policy CR created/updated
  └─> Add finalizer
  └─> Lookup matching Tool CR (by policy.spec.type)
  └─> SetOwnerReference (Tool owns Policy)
  └─> If disabled: Engine.Cleanup, return
  └─> Engine.Validate (spec correctness)
  └─> Engine.Reconcile (create tool-specific resources)
  └─> Engine.GenerateReport (get or build PolicyReport object)
  └─> CreateOrUpdate PolicyReport CR
  └─> If generation changed: reset report to Pending
  └─> Update Policy status (lastReport, observedGeneration)

Policy CR deleted
  └─> Engine.Cleanup (delete Jobs / ConfigMaps / ChaosEngines / Reports)
  └─> Remove finalizer
```

#### PolicyReport Controller

```
PolicyReport CR reconciled
  └─> Skip if phase == Completed
  └─> Lookup parent Policy CR; delete orphan if not found
  └─> ReportHandler.Reconcile
        Trivy:  List VulnerabilityReports from trivy-operator matching scan Jobs
                → Aggregate vulnerability counts, set verdict
        Falco:  Fetch Falco event counts, enrich with pod details
                → Set verdict based on event occurrence
        Litmus: Read ChaosResult resources for each ChaosEngine
                → Determine pass/fail from chaos experiment outcomes
  └─> Update PolicyReport status (phase, summary, details)
```

### Key Interfaces

```go
// ToolService — manages Helm lifecycle of a security tool
type ToolService interface {
    Name() string
    EnsureInstalled(ctx context.Context, tool *eirenyx.Tool) error
    EnsureUninstalled(ctx context.Context, tool *eirenyx.Tool) error
    CheckHealth(ctx context.Context, tool *eirenyx.Tool) (bool, error)
}

// Engine — drives policy reconciliation for a specific tool type
type Engine interface {
    Validate(policy *eirenyx.Policy) error
    Reconcile(ctx context.Context, policy *eirenyx.Policy) error
    Cleanup(ctx context.Context, policy *eirenyx.Policy) error
    GenerateReport(ctx context.Context, policy *eirenyx.Policy) (*eirenyx.PolicyReport, error)
}

// Handler — aggregates tool results into a PolicyReport
type Handler interface {
    Reconcile(ctx context.Context, policyReport *eirenyx.PolicyReport) error
}
```

The `factory.go` module wires these interfaces to their concrete implementations based on the `type` field on `Tool` and `Policy` resources.

---

## Getting Started

### Prerequisites

- Go v1.24+
- Docker 17.03+
- kubectl v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster

### Deploy to cluster

**Build and push the operator image:**

```sh
make docker-build docker-push IMG=<registry>/eirenyx:tag
```

**Install CRDs:**

```sh
make install
```

**Deploy the operator:**

```sh
make deploy IMG=<registry>/eirenyx:tag
```

**Apply sample resources:**

```sh
kubectl apply -k config/samples/
```

### Uninstall

```sh
kubectl delete -k config/samples/
make uninstall
make undeploy
```

### Build a distributable installer

```sh
make build-installer IMG=<registry>/eirenyx:tag
# Produces dist/install.yaml — apply with:
kubectl apply -f dist/install.yaml
```

Run `make help` for all available targets.

---

## Contributing

This project is developed as part of an academic thesis.
Contributions are currently limited to bug reports and design discussions.

More information: [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
