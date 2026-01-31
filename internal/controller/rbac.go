package controller

// -----------------------------------------------------------------------------
// Eirenyx Tool CRDs
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=tools/finalizers,verbs=update

// -----------------------------------------------------------------------------
// Eirenyx Policy CRDs
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=policies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=policies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eirenyx.eirenyx,resources=policies/finalizers,verbs=update

// -----------------------------------------------------------------------------
// Namespaces (cluster-scoped)
// Required for tool installation namespaces (e.g. trivy-system)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create

// -----------------------------------------------------------------------------
// Core Kubernetes resources (used by Helm & engines)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=*
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=*
// +kubebuilder:rbac:groups="",resources=secrets,verbs=*
// +kubebuilder:rbac:groups="",resources=services,verbs=*
// +kubebuilder:rbac:groups="",resources=pods,verbs=*

// -----------------------------------------------------------------------------
// Apps resources (installed by tools)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=*
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=*
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=*
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=*

// -----------------------------------------------------------------------------
// Batch resources (Trivy scan Jobs)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

// -----------------------------------------------------------------------------
// RBAC resources (created by Helm charts)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=*
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=*

// -----------------------------------------------------------------------------
// CRDs (installed by tools like Trivy Operator / Litmus / Falco)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete

// -----------------------------------------------------------------------------
// Trivy Operator CRs
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups=aquasecurity.github.io,resources=*,verbs=*

// -----------------------------------------------------------------------------
// Optional: Events (useful for future diagnostics)
// -----------------------------------------------------------------------------

// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
