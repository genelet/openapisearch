package openapisearch

// Transcript records the authoring conversation and tool observations that led
// to draft artifacts.
type Transcript struct {
	Turns []TranscriptTurn `json:"turns,omitempty"`
}

// TranscriptTurn is one ordered entry in an authoring transcript.
type TranscriptTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Source  string `json:"source,omitempty"`
}

// Diagnostic describes an authoring or validation issue.
type Diagnostic struct {
	Severity    string `json:"severity"`
	Code        string `json:"code,omitempty"`
	Message     string `json:"message"`
	Path        string `json:"path,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

// Slot describes a missing or variable value needed by a draft artifact.
type Slot struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	Source      string `json:"source,omitempty"`
}

// Assumption is a provisional authoring choice that must remain reviewable.
type Assumption struct {
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}

// SymbolicBinding names a runtime-provided capability without resolving it.
type SymbolicBinding struct {
	Name        string `json:"name"`
	Kind        string `json:"kind,omitempty"`
	Source      string `json:"source,omitempty"`
	Description string `json:"description,omitempty"`
}

// ReadinessIssue explains why an artifact or operation needs more review before
// validation, rendering, or execution.
type ReadinessIssue struct {
	Severity    string `json:"severity"`
	Code        string `json:"code,omitempty"`
	Message     string `json:"message"`
	OperationID string `json:"operation_id,omitempty"`
	Path        string `json:"path,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

// QuestionPlan lists clarification questions that would reduce ambiguity.
type QuestionPlan struct {
	Questions []Question `json:"questions,omitempty"`
}

// Question is one caller-facing clarification prompt.
type Question struct {
	Name      string   `json:"name,omitempty"`
	Prompt    string   `json:"prompt"`
	Options   []string `json:"options,omitempty"`
	IssueCode string   `json:"issue_code,omitempty"`
}

// Artifact is a generated draft file or metadata payload.
type Artifact struct {
	Path      string `json:"path"`
	MediaType string `json:"media_type,omitempty"`
	Content   []byte `json:"content"`
}

// ArtifactSet groups draft artifacts with the review metadata used to produce
// them.
type ArtifactSet struct {
	Artifacts        []Artifact          `json:"artifacts,omitempty"`
	Diagnostics      []Diagnostic        `json:"diagnostics,omitempty"`
	Transcript       *Transcript         `json:"transcript,omitempty"`
	Inventory        *OperationInventory `json:"inventory,omitempty"`
	SymbolicBindings []SymbolicBinding   `json:"symbolic_bindings,omitempty"`
	ReadinessIssues  []ReadinessIssue    `json:"readiness_issues,omitempty"`
	Slots            []Slot              `json:"slots,omitempty"`
	Assumptions      []Assumption        `json:"assumptions,omitempty"`
	QuestionPlan     QuestionPlan        `json:"question_plan,omitempty"`
}
