package report

// Local structs modeling the subset of Trivy's `image --format json` output
// the operator cares about. We don't import any types from
// github.com/aquasecurity/trivy directly — Trivy's Go module is large and
// frequently breaks API; modeling only what we render keeps this self-contained.

// TrivyCLIReport is the top-level shape Trivy emits for a single image scan.
type TrivyCLIReport struct {
	ArtifactName string            `json:"ArtifactName,omitempty"`
	ArtifactType string            `json:"ArtifactType,omitempty"`
	Results      []TrivyCLIResult  `json:"Results,omitempty"`
	Metadata     *TrivyCLIMetadata `json:"Metadata,omitempty"`
}

// TrivyCLIMetadata carries OS / image identification. We only use OS for the
// UI's "image" field when ArtifactName is absent.
type TrivyCLIMetadata struct {
	OS *TrivyCLIOS `json:"OS,omitempty"`
}

type TrivyCLIOS struct {
	Family string `json:"Family,omitempty"`
	Name   string `json:"Name,omitempty"`
}

// TrivyCLIResult is one section of a scan (e.g. "debian 12" OS packages, or a
// language ecosystem like "node-pkg"). Trivy emits one Result per scannable
// target inside the image.
type TrivyCLIResult struct {
	Target          string                  `json:"Target,omitempty"`
	Class           string                  `json:"Class,omitempty"` // "os-pkgs", "lang-pkgs", …
	Type            string                  `json:"Type,omitempty"`  // "debian", "alpine", "npm", …
	Vulnerabilities []TrivyCLIVulnerability `json:"Vulnerabilities,omitempty"`
}

// TrivyCLIVulnerability is one finding. The schema is Trivy's CLI output, which
// uses PascalCase field names — not the UI's camelCase.
type TrivyCLIVulnerability struct {
	VulnerabilityID  string   `json:"VulnerabilityID,omitempty"`
	PkgName          string   `json:"PkgName,omitempty"`
	PkgPath          string   `json:"PkgPath,omitempty"`
	InstalledVersion string   `json:"InstalledVersion,omitempty"`
	FixedVersion     string   `json:"FixedVersion,omitempty"`
	Severity         string   `json:"Severity,omitempty"`
	Title            string   `json:"Title,omitempty"`
	Description      string   `json:"Description,omitempty"`
	PrimaryURL       string   `json:"PrimaryURL,omitempty"`
	References       []string `json:"References,omitempty"`
}

// UIVulnerability is the camelCase shape consumed by the eirenyx UI
// (see eirenyx-ui/src/app/core/model/report.model.ts -> TrivyVulnerability).
// We build this from TrivyCLIVulnerability so the UI keeps working unchanged.
type UIVulnerability struct {
	VulnerabilityID  string `json:"vulnerabilityID"`
	Resource         string `json:"resource"`
	InstalledVersion string `json:"installedVersion"`
	FixedVersion     string `json:"fixedVersion,omitempty"`
	Severity         string `json:"severity"`
	Title            string `json:"title,omitempty"`
	Description      string `json:"description,omitempty"`
	PrimaryLink      string `json:"primaryLink,omitempty"`
}

// TrivyDetails is the JSON the operator writes into PolicyReport.status.details
// for Trivy reports. Matches eirenyx-ui's `TrivyReportDetails`.
type TrivyDetails struct {
	// Image is the artifact the user asked us to scan (e.g. "nginx:1.31-perl").
	Image           string            `json:"image,omitempty"`
	Vulnerabilities []UIVulnerability `json:"vulnerabilities,omitempty"`
	// ReportCount is the number of distinct scans that contributed to this
	// PolicyReport. Today it equals len(policy.Spec.Trivy.Scans) minus failures.
	ReportCount int `json:"reportCount,omitempty"`
}

// flattenVulnerabilities walks one TrivyCLIReport's Results and emits the
// UI-shaped vulnerability list. `Resource` is best-effort: PkgName when present,
// PkgPath otherwise (some Trivy classes — e.g. secrets, misconfig — don't have
// either, in which case we leave it blank rather than invent a value).
func flattenVulnerabilities(report *TrivyCLIReport) []UIVulnerability {
	if report == nil {
		return nil
	}
	out := make([]UIVulnerability, 0, 16)
	for _, r := range report.Results {
		for _, v := range r.Vulnerabilities {
			resource := v.PkgName
			if resource == "" {
				resource = v.PkgPath
			}
			out = append(out, UIVulnerability{
				VulnerabilityID:  v.VulnerabilityID,
				Resource:         resource,
				InstalledVersion: v.InstalledVersion,
				FixedVersion:     v.FixedVersion,
				Severity:         v.Severity,
				Title:            v.Title,
				Description:      v.Description,
				PrimaryLink:      v.PrimaryURL,
			})
		}
	}
	return out
}
