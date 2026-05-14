package manager

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

type devWorkspaceStore struct {
	records    map[string]WorkspaceRecord
	orderedIDs []string
}

func (s devWorkspaceStore) GetWorkspace(_ context.Context, id string) (*WorkspaceRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &record, nil
}

func (s devWorkspaceStore) ListWorkspaces(_ context.Context, page, limit int) ([]WorkspaceRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []WorkspaceRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]WorkspaceRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devWorkspaceStore) ListWorkspacesByTenant(_ context.Context, tenantID string, page, limit int) ([]WorkspaceRecord, int, error) {
	filtered := make([]WorkspaceRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	total := len(filtered)
	start := (page - 1) * limit
	if start >= total {
		return []WorkspaceRecord{}, total, nil
	}
	end := min(start+limit, total)
	return filtered[start:end], total, nil
}

func (s *devWorkspaceStore) CreateWorkspace(_ context.Context, workspace WorkspaceRecord) error {
	if _, exists := s.records[workspace.ID]; exists {
		return ErrConflict
	}
	s.records[workspace.ID] = workspace
	s.orderedIDs = append(s.orderedIDs, workspace.ID)
	return nil
}

func (s *devWorkspaceStore) UpdateWorkspace(_ context.Context, id string, fields map[string]string) (*WorkspaceRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["kubernetes_namespace"]; ok {
		rec.KubernetesNamespace = v
	}
	if v, ok := fields["kubernetes_workspace_name"]; ok {
		rec.KubernetesWorkspaceName = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devWorkspaceStore) DeleteWorkspace(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

type devTenantStore struct {
	records    map[string]TenantRecord
	orderedIDs []string
}

func (s devTenantStore) GetTenant(_ context.Context, id string) (*TenantRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &record, nil
}

func (s devTenantStore) ListTenants(_ context.Context, page, limit int) ([]TenantRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []TenantRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]TenantRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s *devTenantStore) CreateTenant(_ context.Context, tenant TenantRecord) error {
	if _, exists := s.records[tenant.ID]; exists {
		return ErrConflict
	}
	s.records[tenant.ID] = tenant
	s.orderedIDs = append(s.orderedIDs, tenant.ID)
	return nil
}

func (s *devTenantStore) UpdateTenant(_ context.Context, id string, fields map[string]string) (*TenantRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["default_region"]; ok {
		rec.DefaultRegion = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devTenantStore) DeleteTenant(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

type devAgentStore struct {
	records    map[string]AgentRecord
	orderedIDs []string
}

func (s devAgentStore) GetAgent(_ context.Context, id string) (*AgentRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &record, nil
}

func (s devAgentStore) ListAgents(_ context.Context, page, limit int) ([]AgentRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []AgentRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]AgentRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devAgentStore) ListAgentsByTenant(_ context.Context, tenantID string, page, limit int) ([]AgentRecord, int, error) {
	filtered := make([]AgentRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateAgents(filtered, page, limit)
}

func (s devAgentStore) ListAgentsByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]AgentRecord, int, error) {
	filtered := make([]AgentRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateAgents(filtered, page, limit)
}

func paginateAgents(records []AgentRecord, page, limit int) ([]AgentRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []AgentRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devAgentStore) CreateAgent(_ context.Context, agent AgentRecord) error {
	if _, exists := s.records[agent.ID]; exists {
		return ErrConflict
	}
	s.records[agent.ID] = agent
	s.orderedIDs = append(s.orderedIDs, agent.ID)
	return nil
}

func (s *devAgentStore) UpdateAgent(_ context.Context, id string, fields map[string]string, spec *AgentSpecData) (*AgentRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["pattern"]; ok {
		rec.Pattern = v
	}
	if v, ok := fields["runtime_engine"]; ok {
		rec.RuntimeEngine = v
	}
	if v, ok := fields["runner_class"]; ok {
		rec.RunnerClass = v
	}
	if v, ok := fields["model_provider"]; ok {
		rec.ModelProvider = v
	}
	if v, ok := fields["model_name"]; ok {
		rec.ModelName = v
	}
	if v, ok := fields["latest_revision"]; ok {
		rec.LatestRevision = v
	}
	if spec != nil {
		rec.Spec = spec
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devAgentStore) UpdateAgentPublish(_ context.Context, id string, fields map[string]string, revision RevisionEntry) (*AgentRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["compile_status"]; ok {
		rec.CompileStatus = v
	}
	if v, ok := fields["latest_revision"]; ok {
		rec.LatestRevision = v
	}
	// Clear compile errors on success, set on error.
	if rec.CompileStatus == "error" {
		rec.CompileErrors = []string{}
	} else {
		rec.CompileErrors = nil
	}
	rec.Revisions = append(rec.Revisions, revision)
	s.records[id] = rec
	return &rec, nil
}

func (s *devAgentStore) DeleteAgent(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

type devEvaluationStore struct {
	records    map[string]EvaluationRecord
	orderedIDs []string
}

func (s devEvaluationStore) GetEvaluation(_ context.Context, id string) (*EvaluationRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &record, nil
}

func (s devEvaluationStore) ListEvaluations(_ context.Context, page, limit int) ([]EvaluationRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []EvaluationRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]EvaluationRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devEvaluationStore) ListEvaluationsByTenant(_ context.Context, tenantID string, page, limit int) ([]EvaluationRecord, int, error) {
	filtered := make([]EvaluationRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateEvaluations(filtered, page, limit)
}

func (s devEvaluationStore) ListEvaluationsByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]EvaluationRecord, int, error) {
	filtered := make([]EvaluationRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateEvaluations(filtered, page, limit)
}

func (s devEvaluationStore) ListEvaluationsByAgent(_ context.Context, agentID string, page, limit int) ([]EvaluationRecord, int, error) {
	filtered := make([]EvaluationRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].AgentID == agentID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateEvaluations(filtered, page, limit)
}

func paginateEvaluations(records []EvaluationRecord, page, limit int) ([]EvaluationRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []EvaluationRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devEvaluationStore) CreateEvaluation(_ context.Context, eval EvaluationRecord) error {
	if _, exists := s.records[eval.ID]; exists {
		return ErrConflict
	}
	s.records[eval.ID] = eval
	s.orderedIDs = append(s.orderedIDs, eval.ID)
	return nil
}

func (s *devEvaluationStore) UpdateEvaluation(_ context.Context, id string, fields map[string]string) (*EvaluationRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["dataset_name"]; ok {
		rec.DatasetName = v
	}
	if v, ok := fields["dataset_revision"]; ok {
		rec.DatasetRevision = v
	}
	if v, ok := fields["baseline_revision"]; ok {
		rec.BaselineRevision = v
	}
	if v, ok := fields["score"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			rec.Score = f
		} else {
			return nil, fmt.Errorf("invalid score %q: %w", v, err)
		}
	}
	if v, ok := fields["gate_passed"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			rec.GatePassed = b
		} else {
			return nil, fmt.Errorf("invalid gate_passed %q: %w", v, err)
		}
	}
	if v, ok := fields["samples_total"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			rec.SamplesTotal = n
		} else {
			return nil, fmt.Errorf("invalid samples_total %q: %w", v, err)
		}
	}
	if v, ok := fields["samples_evaluated"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			rec.SamplesEvaluated = n
		} else {
			return nil, fmt.Errorf("invalid samples_evaluated %q: %w", v, err)
		}
	}
	if v, ok := fields["latest_run_id"]; ok {
		rec.LatestRunID = v
	}
	if v, ok := fields["report_ref"]; ok {
		rec.ReportRef = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devEvaluationStore) DeleteEvaluation(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

type devProviderStore struct {
	records    map[string]ProviderRecord
	orderedIDs []string
}

func (s devProviderStore) GetProvider(_ context.Context, id string) (*ProviderRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &record, nil
}

func (s devProviderStore) ListProviders(_ context.Context, page, limit int) ([]ProviderRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []ProviderRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]ProviderRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devProviderStore) ListProvidersByTenant(_ context.Context, tenantID string, page, limit int) ([]ProviderRecord, int, error) {
	filtered := make([]ProviderRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateProviders(filtered, page, limit)
}

func (s devProviderStore) ListProvidersByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]ProviderRecord, int, error) {
	filtered := make([]ProviderRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateProviders(filtered, page, limit)
}

func paginateProviders(records []ProviderRecord, page, limit int) ([]ProviderRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []ProviderRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devProviderStore) CreateProvider(_ context.Context, provider ProviderRecord) error {
	if _, exists := s.records[provider.ID]; exists {
		return ErrConflict
	}
	s.records[provider.ID] = provider
	s.orderedIDs = append(s.orderedIDs, provider.ID)
	return nil
}

func (s *devProviderStore) UpdateProvider(_ context.Context, id string, fields map[string]string) (*ProviderRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["family"]; ok {
		rec.Family = v
	}
	if v, ok := fields["base_url"]; ok {
		rec.BaseURL = v
	}
	if v, ok := fields["credential_ref"]; ok {
		rec.CredentialRef = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["domestic"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			rec.Domestic = b
		} else {
			return nil, fmt.Errorf("invalid domestic %q: %w", v, err)
		}
	}
	if v, ok := fields["supports_json_schema"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			rec.SupportsJSONSchema = b
		} else {
			return nil, fmt.Errorf("invalid supports_json_schema %q: %w", v, err)
		}
	}
	if v, ok := fields["supports_tool_calling"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			rec.SupportsToolCalling = b
		} else {
			return nil, fmt.Errorf("invalid supports_tool_calling %q: %w", v, err)
		}
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devProviderStore) DeleteProvider(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

type devRunStore struct {
	records    map[string]RunRecord
	orderedIDs []string
}

func (s devRunStore) GetRun(_ context.Context, id string) (*RunRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &record, nil
}

func (s devRunStore) ListRuns(_ context.Context, page, limit int) ([]RunRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []RunRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]RunRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devRunStore) ListRunsByTenant(_ context.Context, tenantID string, page, limit int) ([]RunRecord, int, error) {
	filtered := make([]RunRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateRuns(filtered, page, limit)
}

func (s devRunStore) ListRunsByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]RunRecord, int, error) {
	filtered := make([]RunRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateRuns(filtered, page, limit)
}

func (s devRunStore) ListRunsByAgent(_ context.Context, agentID string, page, limit int) ([]RunRecord, int, error) {
	filtered := make([]RunRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].AgentID == agentID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateRuns(filtered, page, limit)
}

func (s devRunStore) ListRunsByEvaluation(_ context.Context, evaluationID string, page, limit int) ([]RunRecord, int, error) {
	filtered := make([]RunRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].EvaluationID == evaluationID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateRuns(filtered, page, limit)
}

func paginateRuns(records []RunRecord, page, limit int) ([]RunRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []RunRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devRunStore) CreateRun(_ context.Context, run RunRecord) error {
	if _, exists := s.records[run.ID]; exists {
		return ErrConflict
	}
	s.records[run.ID] = run
	s.orderedIDs = append(s.orderedIDs, run.ID)
	return nil
}

func (s *devRunStore) UpdateRun(_ context.Context, id string, fields map[string]string) (*RunRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["started_at"]; ok {
		rec.StartedAt = v
	}
	if v, ok := fields["completed_at"]; ok {
		rec.CompletedAt = v
	}
	if v, ok := fields["summary"]; ok {
		rec.Summary = v
	}
	if v, ok := fields["trace_ref"]; ok {
		rec.TraceRef = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devRunStore) DeleteRun(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// PromptTemplate dev store

type devPromptTemplateStore struct {
	records    map[string]PromptTemplateRecord
	orderedIDs []string
}

func (s devPromptTemplateStore) GetPromptTemplate(_ context.Context, id string) (*PromptTemplateRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devPromptTemplateStore) ListPromptTemplates(_ context.Context, page, limit int) ([]PromptTemplateRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []PromptTemplateRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]PromptTemplateRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devPromptTemplateStore) ListPromptTemplatesByTenant(_ context.Context, tenantID string, page, limit int) ([]PromptTemplateRecord, int, error) {
	filtered := make([]PromptTemplateRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginatePromptTemplates(filtered, page, limit)
}

func (s devPromptTemplateStore) ListPromptTemplatesByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]PromptTemplateRecord, int, error) {
	filtered := make([]PromptTemplateRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginatePromptTemplates(filtered, page, limit)
}

func paginatePromptTemplates(records []PromptTemplateRecord, page, limit int) ([]PromptTemplateRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []PromptTemplateRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devPromptTemplateStore) CreatePromptTemplate(_ context.Context, pt PromptTemplateRecord) error {
	if _, exists := s.records[pt.ID]; exists {
		return ErrConflict
	}
	s.records[pt.ID] = pt
	s.orderedIDs = append(s.orderedIDs, pt.ID)
	return nil
}

func (s *devPromptTemplateStore) UpdatePromptTemplate(_ context.Context, id string, fields map[string]string) (*PromptTemplateRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["template"]; ok {
		rec.Template = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devPromptTemplateStore) DeletePromptTemplate(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// ToolProvider dev store

type devToolProviderStore struct {
	records    map[string]ToolProviderRecord
	orderedIDs []string
}

func (s devToolProviderStore) GetToolProvider(_ context.Context, id string) (*ToolProviderRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devToolProviderStore) ListToolProviders(_ context.Context, page, limit int) ([]ToolProviderRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []ToolProviderRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]ToolProviderRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devToolProviderStore) ListToolProvidersByTenant(_ context.Context, tenantID string, page, limit int) ([]ToolProviderRecord, int, error) {
	filtered := make([]ToolProviderRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateToolProviders(filtered, page, limit)
}

func (s devToolProviderStore) ListToolProvidersByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]ToolProviderRecord, int, error) {
	filtered := make([]ToolProviderRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateToolProviders(filtered, page, limit)
}

func paginateToolProviders(records []ToolProviderRecord, page, limit int) ([]ToolProviderRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []ToolProviderRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devToolProviderStore) CreateToolProvider(_ context.Context, tp ToolProviderRecord) error {
	if _, exists := s.records[tp.ID]; exists {
		return ErrConflict
	}
	s.records[tp.ID] = tp
	s.orderedIDs = append(s.orderedIDs, tp.ID)
	return nil
}

func (s *devToolProviderStore) UpdateToolProvider(_ context.Context, id string, fields map[string]string) (*ToolProviderRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["tool_type"]; ok {
		rec.ToolType = v
	}
	if v, ok := fields["endpoint"]; ok {
		rec.Endpoint = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devToolProviderStore) DeleteToolProvider(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// KnowledgeBase dev store

type devKnowledgeBaseStore struct {
	records    map[string]KnowledgeBaseRecord
	orderedIDs []string
}

func (s devKnowledgeBaseStore) GetKnowledgeBase(_ context.Context, id string) (*KnowledgeBaseRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devKnowledgeBaseStore) ListKnowledgeBases(_ context.Context, page, limit int) ([]KnowledgeBaseRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []KnowledgeBaseRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]KnowledgeBaseRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devKnowledgeBaseStore) ListKnowledgeBasesByTenant(_ context.Context, tenantID string, page, limit int) ([]KnowledgeBaseRecord, int, error) {
	filtered := make([]KnowledgeBaseRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateKnowledgeBases(filtered, page, limit)
}

func (s devKnowledgeBaseStore) ListKnowledgeBasesByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]KnowledgeBaseRecord, int, error) {
	filtered := make([]KnowledgeBaseRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateKnowledgeBases(filtered, page, limit)
}

func paginateKnowledgeBases(records []KnowledgeBaseRecord, page, limit int) ([]KnowledgeBaseRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []KnowledgeBaseRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devKnowledgeBaseStore) CreateKnowledgeBase(_ context.Context, kb KnowledgeBaseRecord) error {
	if _, exists := s.records[kb.ID]; exists {
		return ErrConflict
	}
	s.records[kb.ID] = kb
	s.orderedIDs = append(s.orderedIDs, kb.ID)
	return nil
}

func (s *devKnowledgeBaseStore) UpdateKnowledgeBase(_ context.Context, id string, fields map[string]string) (*KnowledgeBaseRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["source_type"]; ok {
		rec.SourceType = v
	}
	if v, ok := fields["source_ref"]; ok {
		rec.SourceRef = v
	}
	if v, ok := fields["embed_model"]; ok {
		rec.EmbedModel = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devKnowledgeBaseStore) DeleteKnowledgeBase(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// Dataset dev store

type devDatasetStore struct {
	records    map[string]DatasetRecord
	orderedIDs []string
}

func (s devDatasetStore) GetDataset(_ context.Context, id string) (*DatasetRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devDatasetStore) ListDatasets(_ context.Context, page, limit int) ([]DatasetRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []DatasetRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]DatasetRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devDatasetStore) ListDatasetsByTenant(_ context.Context, tenantID string, page, limit int) ([]DatasetRecord, int, error) {
	filtered := make([]DatasetRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateDatasets(filtered, page, limit)
}

func (s devDatasetStore) ListDatasetsByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]DatasetRecord, int, error) {
	filtered := make([]DatasetRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateDatasets(filtered, page, limit)
}

func paginateDatasets(records []DatasetRecord, page, limit int) ([]DatasetRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []DatasetRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devDatasetStore) CreateDataset(_ context.Context, ds DatasetRecord) error {
	if _, exists := s.records[ds.ID]; exists {
		return ErrConflict
	}
	s.records[ds.ID] = ds
	s.orderedIDs = append(s.orderedIDs, ds.ID)
	return nil
}

func (s *devDatasetStore) UpdateDataset(_ context.Context, id string, fields map[string]string) (*DatasetRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["format"]; ok {
		rec.Format = v
	}
	if v, ok := fields["source_ref"]; ok {
		rec.SourceRef = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devDatasetStore) DeleteDataset(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// MCPServer dev store

type devMCPServerStore struct {
	records    map[string]MCPServerRecord
	orderedIDs []string
}

func (s devMCPServerStore) GetMCPServer(_ context.Context, id string) (*MCPServerRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devMCPServerStore) ListMCPServers(_ context.Context, page, limit int) ([]MCPServerRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []MCPServerRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]MCPServerRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devMCPServerStore) ListMCPServersByTenant(_ context.Context, tenantID string, page, limit int) ([]MCPServerRecord, int, error) {
	filtered := make([]MCPServerRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateMCPServers(filtered, page, limit)
}

func (s devMCPServerStore) ListMCPServersByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]MCPServerRecord, int, error) {
	filtered := make([]MCPServerRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateMCPServers(filtered, page, limit)
}

func paginateMCPServers(records []MCPServerRecord, page, limit int) ([]MCPServerRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []MCPServerRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devMCPServerStore) CreateMCPServer(_ context.Context, mcp MCPServerRecord) error {
	if _, exists := s.records[mcp.ID]; exists {
		return ErrConflict
	}
	s.records[mcp.ID] = mcp
	s.orderedIDs = append(s.orderedIDs, mcp.ID)
	return nil
}

func (s *devMCPServerStore) UpdateMCPServer(_ context.Context, id string, fields map[string]string) (*MCPServerRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["endpoint"]; ok {
		rec.Endpoint = v
	}
	if v, ok := fields["transport"]; ok {
		rec.Transport = v
	}
	if v, ok := fields["version"]; ok {
		rec.Version = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devMCPServerStore) DeleteMCPServer(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// AgentPolicy dev store

type devAgentPolicyStore struct {
	records    map[string]AgentPolicyRecord
	orderedIDs []string
}

func (s devAgentPolicyStore) GetAgentPolicy(_ context.Context, id string) (*AgentPolicyRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devAgentPolicyStore) ListAgentPolicies(_ context.Context, page, limit int) ([]AgentPolicyRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []AgentPolicyRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]AgentPolicyRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devAgentPolicyStore) ListAgentPoliciesByTenant(_ context.Context, tenantID string, page, limit int) ([]AgentPolicyRecord, int, error) {
	filtered := make([]AgentPolicyRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateAgentPolicies(filtered, page, limit)
}

func (s devAgentPolicyStore) ListAgentPoliciesByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]AgentPolicyRecord, int, error) {
	filtered := make([]AgentPolicyRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateAgentPolicies(filtered, page, limit)
}

func paginateAgentPolicies(records []AgentPolicyRecord, page, limit int) ([]AgentPolicyRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []AgentPolicyRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devAgentPolicyStore) CreateAgentPolicy(_ context.Context, policy AgentPolicyRecord) error {
	if _, exists := s.records[policy.ID]; exists {
		return ErrConflict
	}
	s.records[policy.ID] = policy
	s.orderedIDs = append(s.orderedIDs, policy.ID)
	return nil
}

func (s *devAgentPolicyStore) UpdateAgentPolicy(_ context.Context, id string, fields map[string]string) (*AgentPolicyRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["policy_type"]; ok {
		rec.PolicyType = v
	}
	if v, ok := fields["enforcement"]; ok {
		rec.Enforcement = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devAgentPolicyStore) DeleteAgentPolicy(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// Skill dev store

type devSkillStore struct {
	records    map[string]SkillRecord
	orderedIDs []string
}

func (s devSkillStore) GetSkill(_ context.Context, id string) (*SkillRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s devSkillStore) ListSkills(_ context.Context, page, limit int) ([]SkillRecord, int, error) {
	total := len(s.records)
	start := (page - 1) * limit
	if start >= total {
		return []SkillRecord{}, total, nil
	}
	end := min(start+limit, total)
	result := make([]SkillRecord, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, s.records[s.orderedIDs[i]])
	}
	return result, total, nil
}

func (s devSkillStore) ListSkillsByTenant(_ context.Context, tenantID string, page, limit int) ([]SkillRecord, int, error) {
	filtered := make([]SkillRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].TenantID == tenantID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateSkills(filtered, page, limit)
}

func (s devSkillStore) ListSkillsByWorkspace(_ context.Context, workspaceID string, page, limit int) ([]SkillRecord, int, error) {
	filtered := make([]SkillRecord, 0)
	for _, id := range s.orderedIDs {
		if s.records[id].WorkspaceID == workspaceID {
			filtered = append(filtered, s.records[id])
		}
	}
	return paginateSkills(filtered, page, limit)
}

func paginateSkills(records []SkillRecord, page, limit int) ([]SkillRecord, int, error) {
	total := len(records)
	start := (page - 1) * limit
	if start >= total {
		return []SkillRecord{}, total, nil
	}
	end := min(start+limit, total)
	return records[start:end], total, nil
}

func (s *devSkillStore) CreateSkill(_ context.Context, skill SkillRecord) error {
	if _, exists := s.records[skill.ID]; exists {
		return ErrConflict
	}
	s.records[skill.ID] = skill
	s.orderedIDs = append(s.orderedIDs, skill.ID)
	return nil
}

func (s *devSkillStore) UpdateSkill(_ context.Context, id string, fields map[string]string) (*SkillRecord, error) {
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["display_name"]; ok {
		rec.DisplayName = v
	}
	if v, ok := fields["description"]; ok {
		rec.Description = v
	}
	if v, ok := fields["status"]; ok {
		rec.Status = v
	}
	if v, ok := fields["skill_type"]; ok {
		rec.SkillType = v
	}
	if v, ok := fields["entrypoint"]; ok {
		rec.Entrypoint = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devSkillStore) DeleteSkill(_ context.Context, id string) error {
	delete(s.records, id)
	for i, oid := range s.orderedIDs {
		if oid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// devUserStore is an in-memory UserStore for development.
type devUserStore struct {
	records    map[string]UserRecord
	orderedIDs []string
	mu         sync.RWMutex
}

func (s *devUserStore) GetUser(_ context.Context, id string) (*UserRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &rec, nil
}

func (s *devUserStore) GetUserByUsername(_ context.Context, username string) (*UserRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, rec := range s.records {
		if rec.Username == username {
			return &rec, nil
		}
	}
	return nil, ErrNotFound
}

func (s *devUserStore) ListUsers(_ context.Context) ([]UserRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]UserRecord, 0, len(s.orderedIDs))
	for _, id := range s.orderedIDs {
		result = append(result, s.records[id])
	}
	return result, nil
}

func (s *devUserStore) CreateUser(_ context.Context, u UserRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[u.ID]; exists {
		return ErrConflict
	}
	s.records[u.ID] = u
	s.orderedIDs = append(s.orderedIDs, u.ID)
	return nil
}

func (s *devUserStore) UpdateUser(_ context.Context, id string, fields map[string]string) (*UserRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	if v, ok := fields["username"]; ok {
		rec.Username = v
	}
	if v, ok := fields["password_hash"]; ok {
		rec.PasswordHash = v
	}
	if v, ok := fields["role"]; ok {
		rec.Role = v
	}
	if v, ok := fields["tenant_id"]; ok {
		rec.TenantID = v
	}
	s.records[id] = rec
	return &rec, nil
}

func (s *devUserStore) DeleteUser(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[id]; !exists {
		return ErrNotFound
	}
	delete(s.records, id)
	for i, uid := range s.orderedIDs {
		if uid == id {
			s.orderedIDs = append(s.orderedIDs[:i], s.orderedIDs[i+1:]...)
			break
		}
	}
	return nil
}

// devSessionStore is an in-memory SessionStore for development.
type devSessionStore struct {
	sessions map[string]Session
	mu       sync.RWMutex
}

func (s *devSessionStore) CreateSession(_ context.Context, sess Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ID] = sess
	return nil
}

func (s *devSessionStore) GetSession(_ context.Context, id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &sess, nil
}

func (s *devSessionStore) DeleteSession(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
	return nil
}

func NewFakeStores() Stores {
	workspaces := &devWorkspaceStore{
		records: map[string]WorkspaceRecord{
			"ws_demo":       {ID: "ws_demo", TenantID: "t_demo", Slug: "demo-ws", DisplayName: "Demo Workspace", Description: "A demo workspace for development", Status: "active", KubernetesNamespace: "demo", KubernetesWorkspaceName: "workspace-demo"},
			"ws_staging":    {ID: "ws_staging", TenantID: "t_demo", Slug: "staging-ws", DisplayName: "Staging Workspace", Status: "active", KubernetesNamespace: "staging"},
			"ws_enterprise": {ID: "ws_enterprise", TenantID: "t_enterprise", Slug: "enterprise-ws", DisplayName: "Enterprise Workspace", Description: "Enterprise customer workspace", Status: "active", KubernetesNamespace: "enterprise", KubernetesWorkspaceName: "workspace-enterprise"},
		},
		orderedIDs: []string{"ws_demo", "ws_staging", "ws_enterprise"},
	}
	tenants := &devTenantStore{
		records: map[string]TenantRecord{
			"t_demo":       {ID: "t_demo", OrganizationID: "org_1", Slug: "demo-tenant", DisplayName: "Demo Tenant", Status: "active", DefaultRegion: "us-east-1"},
			"t_enterprise": {ID: "t_enterprise", OrganizationID: "org_1", Slug: "enterprise-tenant", DisplayName: "Enterprise Tenant", Status: "active", DefaultRegion: "eu-west-1"},
			"t_inactive":   {ID: "t_inactive", OrganizationID: "org_2", Slug: "inactive-tenant", DisplayName: "Inactive Tenant", Status: "inactive"},
		},
		orderedIDs: []string{"t_demo", "t_enterprise", "t_inactive"},
	}
	agents := &devAgentStore{
		records: map[string]AgentRecord{
			"agent_ehs_react": {
				ID: "agent_ehs_react", TenantID: "t_demo", WorkspaceID: "ws_demo", Slug: "ehs-react",
				DisplayName: "EHS ReAct Agent", Description: "Safety incident triage with tools and knowledge", Status: "published",
				Pattern: "react", RuntimeEngine: "eino", RunnerClass: "adk", ModelProvider: "openai", ModelName: "gpt-4.1-mini", LatestRevision: "rev-20260429-001",
			},
			"agent_eval_guard": {
				ID: "agent_eval_guard", TenantID: "t_demo", WorkspaceID: "ws_staging", Slug: "eval-guard",
				DisplayName: "Evaluation Guard", Description: "Release gate evaluator for regression checks", Status: "draft",
				Pattern: "react", RuntimeEngine: "eino", RunnerClass: "adk", ModelProvider: "qwen", ModelName: "qwen-plus", LatestRevision: "rev-20260429-002",
			},
			"agent_enterprise_ops": {
				ID: "agent_enterprise_ops", TenantID: "t_enterprise", WorkspaceID: "ws_enterprise", Slug: "enterprise-ops",
				DisplayName: "Enterprise Ops Agent", Description: "Operations assistant for enterprise workflows", Status: "published",
				Pattern: "react", RuntimeEngine: "eino", RunnerClass: "adk", ModelProvider: "deepseek", ModelName: "deepseek-chat", LatestRevision: "rev-20260429-003",
			},
		},
		orderedIDs: []string{"agent_ehs_react", "agent_eval_guard", "agent_enterprise_ops"},
	}
	evaluations := &devEvaluationStore{
		records: map[string]EvaluationRecord{
			"eval_ehs_regression": {
				ID: "eval_ehs_regression", TenantID: "t_demo", WorkspaceID: "ws_demo", AgentID: "agent_ehs_react",
				Slug: "ehs-regression", DisplayName: "EHS Regression Gate", Description: "Pre-release regression suite for safety incident triage",
				Status: "passed", DatasetName: "ehs-golden-set", DatasetRevision: "dataset-rev-12", BaselineRevision: "rev-20260420-007",
				Score: 0.94, GatePassed: true, SamplesTotal: 128, SamplesEvaluated: 128, LatestRunID: "evalrun-20260429-001", ReportRef: "s3://reports/ehs-regression/latest.json",
			},
			"eval_guardrail_release": {
				ID: "eval_guardrail_release", TenantID: "t_demo", WorkspaceID: "ws_staging", AgentID: "agent_eval_guard",
				Slug: "guardrail-release", DisplayName: "Guardrail Release Check", Description: "Blocking gate for release candidate risk checks",
				Status: "failed", DatasetName: "risk-gate-set", DatasetRevision: "dataset-rev-4", BaselineRevision: "rev-20260421-002",
				Score: 0.72, GatePassed: false, SamplesTotal: 64, SamplesEvaluated: 64, LatestRunID: "evalrun-20260429-002", ReportRef: "s3://reports/guardrail-release/latest.json",
			},
			"eval_enterprise_ops": {
				ID: "eval_enterprise_ops", TenantID: "t_enterprise", WorkspaceID: "ws_enterprise", AgentID: "agent_enterprise_ops",
				Slug: "enterprise-ops-weekly", DisplayName: "Enterprise Ops Weekly", Description: "Weekly regression monitor for enterprise operations",
				Status: "running", DatasetName: "enterprise-ops-set", DatasetRevision: "dataset-rev-9", BaselineRevision: "rev-20260422-001",
				Score: 0.88, GatePassed: true, SamplesTotal: 240, SamplesEvaluated: 180, LatestRunID: "evalrun-20260429-003", ReportRef: "",
			},
		},
		orderedIDs: []string{"eval_ehs_regression", "eval_guardrail_release", "eval_enterprise_ops"},
	}
	providers := &devProviderStore{
		records: map[string]ProviderRecord{
			"provider_qwen_prod": {
				ID: "provider_qwen_prod", TenantID: "t_demo", WorkspaceID: "ws_demo",
				Provider: "qwen", DisplayName: "Qwen Production", Family: "openai-compatible",
				BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", CredentialRef: "secret://demo/qwen-api-key",
				Status: "active", Domestic: true, SupportsJSONSchema: true, SupportsToolCalling: true,
			},
			"provider_deepseek_gate": {
				ID: "provider_deepseek_gate", TenantID: "t_demo", WorkspaceID: "ws_staging",
				Provider: "deepseek", DisplayName: "DeepSeek Release Gate", Family: "openai-compatible",
				BaseURL: "https://api.deepseek.com/v1", CredentialRef: "secret://staging/deepseek-api-key",
				Status: "active", Domestic: true, SupportsJSONSchema: true, SupportsToolCalling: true,
			},
			"provider_openai_fallback": {
				ID: "provider_openai_fallback", TenantID: "t_enterprise", WorkspaceID: "ws_enterprise",
				Provider: "openai", DisplayName: "OpenAI Fallback", Family: "openai-compatible",
				BaseURL: "https://api.openai.com/v1", CredentialRef: "secret://enterprise/openai-api-key",
				Status: "active", SupportsJSONSchema: true, SupportsToolCalling: true,
			},
		},
		orderedIDs: []string{"provider_qwen_prod", "provider_deepseek_gate", "provider_openai_fallback"},
	}
	runs := &devRunStore{
		records: map[string]RunRecord{
			"run_ehs_20260429_001": {
				ID: "run_ehs_20260429_001", TenantID: "t_demo", WorkspaceID: "ws_demo", AgentID: "agent_ehs_react",
				AgentRevision: "rev-20260429-001", Status: "succeeded", RuntimeEngine: "eino", RunnerClass: "adk",
				StartedAt: "2026-04-29T09:10:00Z", CompletedAt: "2026-04-29T09:10:14Z",
				Summary: "inspection complete", TraceRef: "pod/run-ehs-20260429-001",
			},
			"run_guardrail_20260429_002": {
				ID: "run_guardrail_20260429_002", TenantID: "t_demo", WorkspaceID: "ws_staging", AgentID: "agent_eval_guard",
				EvaluationID: "eval_guardrail_release", AgentRevision: "rev-20260429-002", Status: "failed",
				RuntimeEngine: "eino", RunnerClass: "adk", StartedAt: "2026-04-29T10:25:00Z",
				CompletedAt: "2026-04-29T10:25:31Z", Summary: "release gate failed", TraceRef: "pod/run-guardrail-20260429-002",
			},
			"run_enterprise_20260429_003": {
				ID: "run_enterprise_20260429_003", TenantID: "t_enterprise", WorkspaceID: "ws_enterprise", AgentID: "agent_enterprise_ops",
				EvaluationID: "eval_enterprise_ops", AgentRevision: "rev-20260429-003", Status: "running",
				RuntimeEngine: "eino", RunnerClass: "adk", StartedAt: "2026-04-29T11:40:00Z",
				Summary: "weekly regression in progress", TraceRef: "pod/run-enterprise-20260429-003",
			},
		},
		orderedIDs: []string{"run_ehs_20260429_001", "run_guardrail_20260429_002", "run_enterprise_20260429_003"},
	}
	promptTemplates := &devPromptTemplateStore{
		records:    map[string]PromptTemplateRecord{},
		orderedIDs: []string{},
	}
	toolProviders := &devToolProviderStore{
		records:    map[string]ToolProviderRecord{},
		orderedIDs: []string{},
	}
	knowledgeBases := &devKnowledgeBaseStore{
		records:    map[string]KnowledgeBaseRecord{},
		orderedIDs: []string{},
	}
	datasets := &devDatasetStore{
		records:    map[string]DatasetRecord{},
		orderedIDs: []string{},
	}
	mcpServers := &devMCPServerStore{
		records:    map[string]MCPServerRecord{},
		orderedIDs: []string{},
	}
	agentPolicies := &devAgentPolicyStore{
		records:    map[string]AgentPolicyRecord{},
		orderedIDs: []string{},
	}
	skills := &devSkillStore{
		records:    map[string]SkillRecord{},
		orderedIDs: []string{},
	}
	users := &devUserStore{
		records:    map[string]UserRecord{},
		orderedIDs: []string{},
	}
	sessions := &devSessionStore{
		sessions: map[string]Session{},
	}
	return Stores{
		Workspaces:      workspaces,
		Tenants:         tenants,
		Agents:          agents,
		Evaluations:     evaluations,
		Providers:       providers,
		Runs:            runs,
		PromptTemplates: promptTemplates,
		ToolProviders:   toolProviders,
		KnowledgeBases:  knowledgeBases,
		Datasets:        datasets,
		MCPServers:      mcpServers,
		AgentPolicies:   agentPolicies,
		Skills:          skills,
		Users:           users,
		Sessions:        sessions,
	}
}
