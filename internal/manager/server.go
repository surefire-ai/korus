package manager

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/surefire-ai/korus/internal/compiler"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	Config Config
	Stores Stores
	Syncer CRDSyncer
}

type storeAwareSyncer interface {
	SetStores(*Stores)
}

type InfoResponse struct {
	Component          string `json:"component"`
	Mode               string `json:"mode"`
	DatabaseConfigured bool   `json:"databaseConfigured"`
	DatabaseDriver     string `json:"databaseDriver,omitempty"`
	DatabaseStatus     string `json:"databaseStatus"`
	MigrateOnStart     bool   `json:"migrateOnStart"`
}

type WorkspaceResponse struct {
	ID                      string `json:"id"`
	TenantID                string `json:"tenantId"`
	Slug                    string `json:"slug"`
	DisplayName             string `json:"displayName"`
	Description             string `json:"description,omitempty"`
	Status                  string `json:"status"`
	KubernetesNamespace     string `json:"kubernetesNamespace,omitempty"`
	KubernetesWorkspaceName string `json:"kubernetesWorkspaceName,omitempty"`
}

type TenantResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Slug           string `json:"slug"`
	DisplayName    string `json:"displayName"`
	Status         string `json:"status"`
	DefaultRegion  string `json:"defaultRegion,omitempty"`
}

type AgentResponse struct {
	ID             string          `json:"id"`
	TenantID       string          `json:"tenantId"`
	WorkspaceID    string          `json:"workspaceId"`
	Slug           string          `json:"slug"`
	DisplayName    string          `json:"displayName"`
	Description    string          `json:"description,omitempty"`
	Status         string          `json:"status"`
	Pattern        string          `json:"pattern"`
	RuntimeEngine  string          `json:"runtimeEngine"`
	RunnerClass    string          `json:"runnerClass"`
	ModelProvider  string          `json:"modelProvider,omitempty"`
	ModelName      string          `json:"modelName,omitempty"`
	LatestRevision string          `json:"latestRevision,omitempty"`
	CompileStatus  string          `json:"compileStatus,omitempty"`
	CompileErrors  []string        `json:"compileErrors,omitempty"`
	Revisions      []RevisionEntry `json:"revisions,omitempty"`
	Spec           *AgentSpecData  `json:"spec,omitempty"`
}

type EvaluationResponse struct {
	ID               string  `json:"id"`
	TenantID         string  `json:"tenantId"`
	WorkspaceID      string  `json:"workspaceId"`
	AgentID          string  `json:"agentId"`
	Slug             string  `json:"slug"`
	DisplayName      string  `json:"displayName"`
	Description      string  `json:"description,omitempty"`
	Status           string  `json:"status"`
	DatasetName      string  `json:"datasetName"`
	DatasetRevision  string  `json:"datasetRevision,omitempty"`
	BaselineRevision string  `json:"baselineRevision,omitempty"`
	Score            float64 `json:"score"`
	GatePassed       bool    `json:"gatePassed"`
	SamplesTotal     int     `json:"samplesTotal"`
	SamplesEvaluated int     `json:"samplesEvaluated"`
	LatestRunID      string  `json:"latestRunId,omitempty"`
	ReportRef        string  `json:"reportRef,omitempty"`
}

type ProviderResponse struct {
	ID                  string `json:"id"`
	TenantID            string `json:"tenantId"`
	WorkspaceID         string `json:"workspaceId,omitempty"`
	Provider            string `json:"provider"`
	DisplayName         string `json:"displayName"`
	Family              string `json:"family"`
	BaseURL             string `json:"baseUrl,omitempty"`
	CredentialRef       string `json:"credentialRef,omitempty"`
	Status              string `json:"status"`
	Domestic            bool   `json:"domestic"`
	SupportsJSONSchema  bool   `json:"supportsJsonSchema"`
	SupportsToolCalling bool   `json:"supportsToolCalling"`
}

type RunResponse struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenantId"`
	WorkspaceID   string `json:"workspaceId"`
	AgentID       string `json:"agentId"`
	EvaluationID  string `json:"evaluationId,omitempty"`
	AgentRevision string `json:"agentRevision,omitempty"`
	Status        string `json:"status"`
	RuntimeEngine string `json:"runtimeEngine"`
	RunnerClass   string `json:"runnerClass"`
	StartedAt     string `json:"startedAt,omitempty"`
	CompletedAt   string `json:"completedAt,omitempty"`
	Summary       string `json:"summary,omitempty"`
	TraceRef      string `json:"traceRef,omitempty"`
}

type PaginatedWorkspacesResponse struct {
	Workspaces []WorkspaceResponse `json:"workspaces"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	Total      int                 `json:"total"`
}

type PaginatedTenantsResponse struct {
	Tenants []TenantResponse `json:"tenants"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
	Total   int              `json:"total"`
}

type PaginatedAgentsResponse struct {
	Agents []AgentResponse `json:"agents"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
	Total  int             `json:"total"`
}

type PaginatedEvaluationsResponse struct {
	Evaluations []EvaluationResponse `json:"evaluations"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
	Total       int                  `json:"total"`
}

type PaginatedProvidersResponse struct {
	Providers []ProviderResponse `json:"providers"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
	Total     int                `json:"total"`
}

type PaginatedRunsResponse struct {
	Runs  []RunResponse `json:"runs"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Total int           `json:"total"`
}

type CreateWorkspaceRequest struct {
	ID                      string `json:"id"`
	TenantID                string `json:"tenantId"`
	Slug                    string `json:"slug"`
	DisplayName             string `json:"displayName"`
	Description             string `json:"description,omitempty"`
	Status                  string `json:"status,omitempty"`
	KubernetesNamespace     string `json:"kubernetesNamespace,omitempty"`
	KubernetesWorkspaceName string `json:"kubernetesWorkspaceName,omitempty"`
}

type UpdateWorkspaceRequest struct {
	DisplayName             *string `json:"displayName,omitempty"`
	Description             *string `json:"description,omitempty"`
	Status                  *string `json:"status,omitempty"`
	KubernetesNamespace     *string `json:"kubernetesNamespace,omitempty"`
	KubernetesWorkspaceName *string `json:"kubernetesWorkspaceName,omitempty"`
}

type CreateTenantRequest struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Slug           string `json:"slug"`
	DisplayName    string `json:"displayName"`
	Status         string `json:"status,omitempty"`
	DefaultRegion  string `json:"defaultRegion,omitempty"`
}

type UpdateTenantRequest struct {
	DisplayName   *string `json:"displayName,omitempty"`
	Status        *string `json:"status,omitempty"`
	DefaultRegion *string `json:"defaultRegion,omitempty"`
}

type CreateAgentRequest struct {
	ID            string         `json:"id"`
	TenantID      string         `json:"tenantId"`
	WorkspaceID   string         `json:"workspaceId"`
	Slug          string         `json:"slug"`
	DisplayName   string         `json:"displayName"`
	Description   string         `json:"description,omitempty"`
	Status        string         `json:"status,omitempty"`
	Pattern       string         `json:"pattern,omitempty"`
	RuntimeEngine string         `json:"runtimeEngine,omitempty"`
	RunnerClass   string         `json:"runnerClass,omitempty"`
	ModelProvider string         `json:"modelProvider,omitempty"`
	ModelName     string         `json:"modelName,omitempty"`
	Spec          *AgentSpecData `json:"spec,omitempty"`
}

type UpdateAgentRequest struct {
	DisplayName   *string        `json:"displayName,omitempty"`
	Description   *string        `json:"description,omitempty"`
	Status        *string        `json:"status,omitempty"`
	Pattern       *string        `json:"pattern,omitempty"`
	RuntimeEngine *string        `json:"runtimeEngine,omitempty"`
	RunnerClass   *string        `json:"runnerClass,omitempty"`
	ModelProvider *string        `json:"modelProvider,omitempty"`
	ModelName     *string        `json:"modelName,omitempty"`
	Spec          *AgentSpecData `json:"spec,omitempty"`
}

type CreateEvaluationRequest struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenantId"`
	WorkspaceID      string `json:"workspaceId"`
	AgentID          string `json:"agentId"`
	Slug             string `json:"slug"`
	DisplayName      string `json:"displayName"`
	Description      string `json:"description,omitempty"`
	Status           string `json:"status,omitempty"`
	DatasetName      string `json:"datasetName,omitempty"`
	DatasetRevision  string `json:"datasetRevision,omitempty"`
	BaselineRevision string `json:"baselineRevision,omitempty"`
}

type UpdateEvaluationRequest struct {
	DisplayName      *string  `json:"displayName,omitempty"`
	Description      *string  `json:"description,omitempty"`
	Status           *string  `json:"status,omitempty"`
	DatasetName      *string  `json:"datasetName,omitempty"`
	DatasetRevision  *string  `json:"datasetRevision,omitempty"`
	BaselineRevision *string  `json:"baselineRevision,omitempty"`
	Score            *float64 `json:"score,omitempty"`
	GatePassed       *bool    `json:"gatePassed,omitempty"`
	SamplesTotal     *int     `json:"samplesTotal,omitempty"`
	SamplesEvaluated *int     `json:"samplesEvaluated,omitempty"`
	LatestRunID      *string  `json:"latestRunId,omitempty"`
	ReportRef        *string  `json:"reportRef,omitempty"`
}

type CreateProviderRequest struct {
	ID                  string `json:"id"`
	TenantID            string `json:"tenantId"`
	WorkspaceID         string `json:"workspaceId,omitempty"`
	Provider            string `json:"provider"`
	DisplayName         string `json:"displayName"`
	Family              string `json:"family,omitempty"`
	BaseURL             string `json:"baseUrl,omitempty"`
	CredentialRef       string `json:"credentialRef,omitempty"`
	Status              string `json:"status,omitempty"`
	Domestic            bool   `json:"domestic,omitempty"`
	SupportsJSONSchema  bool   `json:"supportsJsonSchema,omitempty"`
	SupportsToolCalling bool   `json:"supportsToolCalling,omitempty"`
}

type UpdateProviderRequest struct {
	DisplayName         *string `json:"displayName,omitempty"`
	Family              *string `json:"family,omitempty"`
	BaseURL             *string `json:"baseUrl,omitempty"`
	CredentialRef       *string `json:"credentialRef,omitempty"`
	Status              *string `json:"status,omitempty"`
	Domestic            *bool   `json:"domestic,omitempty"`
	SupportsJSONSchema  *bool   `json:"supportsJsonSchema,omitempty"`
	SupportsToolCalling *bool   `json:"supportsToolCalling,omitempty"`
}

type CreateRunRequest struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenantId"`
	WorkspaceID   string `json:"workspaceId"`
	AgentID       string `json:"agentId"`
	EvaluationID  string `json:"evaluationId,omitempty"`
	AgentRevision string `json:"agentRevision,omitempty"`
	Status        string `json:"status,omitempty"`
	RuntimeEngine string `json:"runtimeEngine,omitempty"`
	RunnerClass   string `json:"runnerClass,omitempty"`
}

type UpdateRunRequest struct {
	Status      *string `json:"status,omitempty"`
	StartedAt   *string `json:"startedAt,omitempty"`
	CompletedAt *string `json:"completedAt,omitempty"`
	Summary     *string `json:"summary,omitempty"`
	TraceRef    *string `json:"traceRef,omitempty"`
}

// PromptTemplate types

type PromptTemplateResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	WorkspaceID string   `json:"workspaceId"`
	Slug        string   `json:"slug"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Template    string   `json:"template,omitempty"`
	Variables   []string `json:"variables,omitempty"`
}

type PaginatedPromptTemplatesResponse struct {
	Templates []PromptTemplateResponse `json:"templates"`
	Page      int                      `json:"page"`
	Limit     int                      `json:"limit"`
	Total     int                      `json:"total"`
}

type CreatePromptTemplateRequest struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	WorkspaceID string   `json:"workspaceId"`
	Slug        string   `json:"slug"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Template    string   `json:"template,omitempty"`
	Variables   []string `json:"variables,omitempty"`
}

type UpdatePromptTemplateRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Template    *string `json:"template,omitempty"`
}

// ToolProvider types

type ToolProviderResponse struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenantId"`
	WorkspaceID string                 `json:"workspaceId"`
	Slug        string                 `json:"slug"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	ToolType    string                 `json:"toolType,omitempty"`
	Endpoint    string                 `json:"endpoint,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type PaginatedToolProvidersResponse struct {
	Tools []ToolProviderResponse `json:"tools"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
	Total int                    `json:"total"`
}

type CreateToolProviderRequest struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenantId"`
	WorkspaceID string                 `json:"workspaceId"`
	Slug        string                 `json:"slug"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	ToolType    string                 `json:"toolType,omitempty"`
	Endpoint    string                 `json:"endpoint,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type UpdateToolProviderRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	ToolType    *string `json:"toolType,omitempty"`
	Endpoint    *string `json:"endpoint,omitempty"`
}

// KnowledgeBase types

type KnowledgeBaseResponse struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenantId"`
	WorkspaceID  string `json:"workspaceId"`
	Slug         string `json:"slug"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status"`
	SourceType   string `json:"sourceType,omitempty"`
	SourceRef    string `json:"sourceRef,omitempty"`
	EmbedModel   string `json:"embedModel,omitempty"`
	ChunkSize    int    `json:"chunkSize,omitempty"`
	ChunkOverlap int    `json:"chunkOverlap,omitempty"`
}

type PaginatedKnowledgeBasesResponse struct {
	Bases []KnowledgeBaseResponse `json:"bases"`
	Page  int                     `json:"page"`
	Limit int                     `json:"limit"`
	Total int                     `json:"total"`
}

type CreateKnowledgeBaseRequest struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenantId"`
	WorkspaceID  string `json:"workspaceId"`
	Slug         string `json:"slug"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status,omitempty"`
	SourceType   string `json:"sourceType,omitempty"`
	SourceRef    string `json:"sourceRef,omitempty"`
	EmbedModel   string `json:"embedModel,omitempty"`
	ChunkSize    int    `json:"chunkSize,omitempty"`
	ChunkOverlap int    `json:"chunkOverlap,omitempty"`
}

type UpdateKnowledgeBaseRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	SourceType  *string `json:"sourceType,omitempty"`
	SourceRef   *string `json:"sourceRef,omitempty"`
	EmbedModel  *string `json:"embedModel,omitempty"`
}

// Dataset types

type DatasetResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	WorkspaceID string   `json:"workspaceId"`
	Slug        string   `json:"slug"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Format      string   `json:"format,omitempty"`
	SourceRef   string   `json:"sourceRef,omitempty"`
	RowCount    int      `json:"rowCount,omitempty"`
	Columns     []string `json:"columns,omitempty"`
}

type PaginatedDatasetsResponse struct {
	Datasets []DatasetResponse `json:"datasets"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
	Total    int               `json:"total"`
}

type CreateDatasetRequest struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	WorkspaceID string   `json:"workspaceId"`
	Slug        string   `json:"slug"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	Format      string   `json:"format,omitempty"`
	SourceRef   string   `json:"sourceRef,omitempty"`
	RowCount    int      `json:"rowCount,omitempty"`
	Columns     []string `json:"columns,omitempty"`
}

type UpdateDatasetRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Format      *string `json:"format,omitempty"`
	SourceRef   *string `json:"sourceRef,omitempty"`
}

// MCPServer types

type MCPServerResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	WorkspaceID string `json:"workspaceId"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	Endpoint    string `json:"endpoint,omitempty"`
	Transport   string `json:"transport,omitempty"`
	Version     string `json:"version,omitempty"`
}

type PaginatedMCPServersResponse struct {
	Servers []MCPServerResponse `json:"servers"`
	Page    int                 `json:"page"`
	Limit   int                 `json:"limit"`
	Total   int                 `json:"total"`
}

type CreateMCPServerRequest struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenantId"`
	WorkspaceID string `json:"workspaceId"`
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	Transport   string `json:"transport,omitempty"`
	Version     string `json:"version,omitempty"`
}

type UpdateMCPServerRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	Endpoint    *string `json:"endpoint,omitempty"`
	Transport   *string `json:"transport,omitempty"`
	Version     *string `json:"version,omitempty"`
}

// AgentPolicy types

type AgentPolicyResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	WorkspaceID string   `json:"workspaceId"`
	Slug        string   `json:"slug"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	PolicyType  string   `json:"policyType,omitempty"`
	Rules       []string `json:"rules,omitempty"`
	Enforcement string   `json:"enforcement,omitempty"`
}

type PaginatedAgentPoliciesResponse struct {
	Policies []AgentPolicyResponse `json:"policies"`
	Page     int                   `json:"page"`
	Limit    int                   `json:"limit"`
	Total    int                   `json:"total"`
}

type CreateAgentPolicyRequest struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenantId"`
	WorkspaceID string   `json:"workspaceId"`
	Slug        string   `json:"slug"`
	DisplayName string   `json:"displayName"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status,omitempty"`
	PolicyType  string   `json:"policyType,omitempty"`
	Rules       []string `json:"rules,omitempty"`
	Enforcement string   `json:"enforcement,omitempty"`
}

type UpdateAgentPolicyRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	PolicyType  *string `json:"policyType,omitempty"`
	Enforcement *string `json:"enforcement,omitempty"`
}

// Skill types

type SkillResponse struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenantId"`
	WorkspaceID string                 `json:"workspaceId"`
	Slug        string                 `json:"slug"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	SkillType   string                 `json:"skillType,omitempty"`
	Entrypoint  string                 `json:"entrypoint,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type PaginatedSkillsResponse struct {
	Skills []SkillResponse `json:"skills"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
	Total  int             `json:"total"`
}

type CreateSkillRequest struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenantId"`
	WorkspaceID string                 `json:"workspaceId"`
	Slug        string                 `json:"slug"`
	DisplayName string                 `json:"displayName"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	SkillType   string                 `json:"skillType,omitempty"`
	Entrypoint  string                 `json:"entrypoint,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type UpdateSkillRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
	SkillType   *string `json:"skillType,omitempty"`
	Entrypoint  *string `json:"entrypoint,omitempty"`
}

// Auth types

type contextKey string

const userContextKey contextKey = "user"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TenantID string `json:"tenantId,omitempty"`
}

type SetupResponse struct {
	Password string `json:"password"`
}

func (s Server) syncer() CRDSyncer {
	if s.Syncer == nil {
		return NoopCRDSyncer{}
	}
	return s.Syncer
}

func (s Server) Start(ctx context.Context) error {
	config := s.Config.normalized()
	database, err := OpenDatabase(ctx, config)
	if err != nil {
		return err
	}
	defer func() {
		_ = database.Close()
	}()
	if database != nil && config.AutoMigrate {
		if _, err := database.ApplyBuiltInMigrations(ctx); err != nil {
			return err
		}
	}
	if database != nil && s.Stores.Workspaces == nil {
		s.Stores = NewSQLStores(database.DB)
	}
	if syncer, ok := s.Syncer.(storeAwareSyncer); ok {
		syncer.SetStores(&s.Stores)
	}

	// First-run admin setup
	if s.Stores.Users != nil {
		users, _ := s.Stores.Users.ListUsers(ctx)
		if len(users) == 0 {
			password := generateRandomString(16)
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash admin password: %w", err)
			}
			if err := s.Stores.Users.CreateUser(ctx, UserRecord{
				ID:           "admin",
				Username:     "admin",
				PasswordHash: string(hash),
				Role:         "admin",
				TenantID:     "",
				CreatedAt:    time.Now().UTC().Format(time.RFC3339),
			}); err != nil {
				return fmt.Errorf("failed to create admin user: %w", err)
			}
			log.Printf("══════════════════════════════════════════════════")
			log.Printf("  ADMIN PASSWORD: %s", password)
			log.Printf("  Username: admin")
			log.Printf("══════════════════════════════════════════════════")
		}
	}

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		err := <-errCh
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/api/v1/info", s.handleInfo)
	mux.HandleFunc("/api/v1/auth/login", s.handleAuthLogin)
	mux.HandleFunc("/api/v1/auth/logout", s.handleAuthLogout)
	mux.HandleFunc("/api/v1/auth/me", s.handleAuthMe)
	mux.HandleFunc("/api/v1/auth/setup", s.handleAuthSetup)
	mux.HandleFunc("/api/v1/workspaces/", s.handleWorkspace)
	mux.HandleFunc("/api/v1/tenants/", s.handleTenant)
	mux.HandleFunc("/api/v1/agents/", s.handleAgent)
	mux.HandleFunc("/api/v1/evaluations/", s.handleEvaluation)
	mux.HandleFunc("/api/v1/providers/", s.handleProvider)
	mux.HandleFunc("/api/v1/runs/", s.handleRun)
	mux.HandleFunc("/api/v1/prompt-templates/", s.handlePromptTemplate)
	mux.HandleFunc("/api/v1/tool-providers/", s.handleToolProvider)
	mux.HandleFunc("/api/v1/knowledge-bases/", s.handleKnowledgeBase)
	mux.HandleFunc("/api/v1/datasets/", s.handleDataset)
	mux.HandleFunc("/api/v1/mcp-servers/", s.handleMCPServer)
	mux.HandleFunc("/api/v1/agent-policies/", s.handleAgentPolicy)
	mux.HandleFunc("/api/v1/skills/", s.handleSkill)
	return corsMiddleware(s.authMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public routes that don't require auth
		path := r.URL.Path
		if path == "/healthz" || path == "/readyz" || path == "/api/v1/info" ||
			path == "/api/v1/auth/login" || path == "/api/v1/auth/logout" || path == "/api/v1/auth/setup" {
			next.ServeHTTP(w, r)
			return
		}

		if s.Stores.Sessions == nil {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		sess, err := s.Stores.Sessions.GetSession(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid session")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)[:n]
}

// Auth handlers

func (s Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method must be POST")
		return
	}
	if s.Stores.Users == nil || s.Stores.Sessions == nil {
		writeError(w, http.StatusServiceUnavailable, "auth not configured")
		return
	}
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	user, err := s.Stores.Users.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	sessionID := generateRandomString(32)
	sess := Session{
		ID:        sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TenantID:  user.TenantID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.Stores.Sessions.CreateSession(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})
	writeJSON(w, http.StatusOK, UserResponse{
		ID: user.ID, Username: user.Username, Role: user.Role, TenantID: user.TenantID,
	})
}

func (s Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method must be POST")
		return
	}
	cookie, err := r.Cookie("session_id")
	if err == nil && s.Stores.Sessions != nil {
		_ = s.Stores.Sessions.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method must be GET")
		return
	}
	sess, ok := r.Context().Value(userContextKey).(*Session)
	if !ok || sess == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, UserResponse{
		ID: sess.UserID, Username: sess.Username, Role: sess.Role, TenantID: sess.TenantID,
	})
}

func (s Server) handleAuthSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method must be POST")
		return
	}
	if s.Stores.Users == nil {
		writeError(w, http.StatusServiceUnavailable, "auth not configured")
		return
	}
	users, _ := s.Stores.Users.ListUsers(r.Context())
	if len(users) > 0 {
		writeError(w, http.StatusConflict, "admin already exists")
		return
	}
	password := generateRandomString(16)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	if err := s.Stores.Users.CreateUser(r.Context(), UserRecord{
		ID:           "admin",
		Username:     "admin",
		PasswordHash: string(hash),
		Role:         "admin",
		TenantID:     "",
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create admin user")
		return
	}
	writeJSON(w, http.StatusCreated, SetupResponse{Password: password})
}

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method must be GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method must be GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method must be GET")
		return
	}
	config := s.Config.normalized()
	databaseStatus := "not_configured"
	if config.DatabaseURL != "" {
		databaseStatus = "configured"
	}
	writeJSON(w, http.StatusOK, InfoResponse{
		Component:          "manager",
		Mode:               config.Mode,
		DatabaseConfigured: config.DatabaseURL != "",
		DatabaseDriver:     config.DatabaseDriver,
		DatabaseStatus:     databaseStatus,
		MigrateOnStart:     config.AutoMigrate,
	})
}

func (s Server) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Workspaces == nil {
		writeError(w, http.StatusServiceUnavailable, "workspace store is not configured")
		return
	}
	workspaceID := strings.TrimPrefix(r.URL.Path, "/api/v1/workspaces/")
	workspaceID = strings.TrimSpace(workspaceID)

	if workspaceID == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListWorkspaces(w, r)
		case http.MethodPost:
			s.handleCreateWorkspace(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(workspaceID, "/") {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetWorkspace(w, r, workspaceID)
	case http.MethodPatch:
		s.handleUpdateWorkspace(w, r, workspaceID)
	case http.MethodDelete:
		s.handleDeleteWorkspace(w, r, workspaceID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	var records []WorkspaceRecord
	var total int
	var err error
	if tenantID != "" {
		records, total, err = s.Stores.Workspaces.ListWorkspacesByTenant(r.Context(), tenantID, page, limit)
	} else {
		records, total, err = s.Stores.Workspaces.ListWorkspaces(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workspaces")
		return
	}
	workspaces := make([]WorkspaceResponse, 0, len(records))
	for _, rec := range records {
		workspaces = append(workspaces, workspaceResponseFromRecord(rec))
	}
	writeJSON(w, http.StatusOK, PaginatedWorkspacesResponse{
		Workspaces: workspaces,
		Page:       page,
		Limit:      limit,
		Total:      total,
	})
}

func (s Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkspaceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, slug, and displayName are required")
		return
	}
	record := WorkspaceRecord{
		ID:                      req.ID,
		TenantID:                req.TenantID,
		Slug:                    req.Slug,
		DisplayName:             req.DisplayName,
		Description:             req.Description,
		Status:                  req.Status,
		KubernetesNamespace:     req.KubernetesNamespace,
		KubernetesWorkspaceName: req.KubernetesWorkspaceName,
	}
	if record.Status == "" {
		record.Status = "active"
	}
	if err := s.Stores.Workspaces.CreateWorkspace(r.Context(), record); err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "workspace already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create workspace")
		return
	}
	if err := s.syncer().SyncWorkspace(r.Context(), record); err != nil {
		log.Printf("syncer: failed to sync workspace %s: %v", record.ID, err)
	}
	writeJSON(w, http.StatusCreated, workspaceResponseFromRecord(record))
}

func (s Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	workspace, err := s.Stores.Workspaces.GetWorkspace(r.Context(), workspaceID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read workspace")
		return
	}
	writeJSON(w, http.StatusOK, workspaceResponseFromRecord(*workspace))
}

func (s Server) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	var req UpdateWorkspaceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.KubernetesNamespace != nil {
		fields["kubernetes_namespace"] = *req.KubernetesNamespace
	}
	if req.KubernetesWorkspaceName != nil {
		fields["kubernetes_workspace_name"] = *req.KubernetesWorkspaceName
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "at least one updatable field must be provided")
		return
	}
	updated, err := s.Stores.Workspaces.UpdateWorkspace(r.Context(), workspaceID, fields)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update workspace")
		return
	}
	if err := s.syncer().SyncWorkspace(r.Context(), *updated); err != nil {
		log.Printf("syncer: failed to sync workspace %s: %v", updated.ID, err)
	}
	writeJSON(w, http.StatusOK, workspaceResponseFromRecord(*updated))
}

func (s Server) handleDeleteWorkspace(w http.ResponseWriter, r *http.Request, workspaceID string) {
	workspace, err := s.Stores.Workspaces.GetWorkspace(r.Context(), workspaceID)
	if errors.Is(err, ErrNotFound) {
		if err := s.Stores.Workspaces.DeleteWorkspace(r.Context(), workspaceID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete workspace")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read workspace")
		return
	}
	if err := s.Stores.Workspaces.DeleteWorkspace(r.Context(), workspaceID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete workspace")
		return
	}
	if err := s.syncer().DeleteWorkspace(r.Context(), *workspace); err != nil {
		log.Printf("syncer: failed to delete workspace %s: %v", workspaceID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleTenant(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Tenants == nil {
		writeError(w, http.StatusServiceUnavailable, "tenant store is not configured")
		return
	}
	tenantID := strings.TrimPrefix(r.URL.Path, "/api/v1/tenants/")
	tenantID = strings.TrimSpace(tenantID)

	if tenantID == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListTenants(w, r)
		case http.MethodPost:
			s.handleCreateTenant(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(tenantID, "/") {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetTenant(w, r, tenantID)
	case http.MethodPatch:
		s.handleUpdateTenant(w, r, tenantID)
	case http.MethodDelete:
		s.handleDeleteTenant(w, r, tenantID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleAgent(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Agents == nil {
		writeError(w, http.StatusServiceUnavailable, "agent store is not configured")
		return
	}
	agentID := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")
	agentID = strings.TrimSpace(agentID)

	if agentID == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListAgents(w, r)
		case http.MethodPost:
			s.handleCreateAgent(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(agentID, "/") {
		// Sub-routes: /api/v1/agents/{id}/compile, /api/v1/agents/{id}/publish
		if strings.HasSuffix(agentID, "/compile") {
			agentID = strings.TrimSuffix(agentID, "/compile")
			if r.Method == http.MethodPost {
				s.handleCompileAgent(w, r, agentID)
				return
			}
			writeError(w, http.StatusMethodNotAllowed, "method must be POST")
			return
		}
		if strings.HasSuffix(agentID, "/publish") {
			agentID = strings.TrimSuffix(agentID, "/publish")
			if r.Method == http.MethodPost {
				s.handlePublishAgent(w, r, agentID)
				return
			}
			writeError(w, http.StatusMethodNotAllowed, "method must be POST")
			return
		}
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetAgent(w, r, agentID)
	case http.MethodPatch:
		s.handleUpdateAgent(w, r, agentID)
	case http.MethodDelete:
		s.handleDeleteAgent(w, r, agentID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleEvaluation(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Evaluations == nil {
		writeError(w, http.StatusServiceUnavailable, "evaluation store is not configured")
		return
	}
	evaluationID := strings.TrimPrefix(r.URL.Path, "/api/v1/evaluations/")
	evaluationID = strings.TrimSpace(evaluationID)

	if evaluationID == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListEvaluations(w, r)
		case http.MethodPost:
			s.handleCreateEvaluation(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(evaluationID, "/") {
		writeError(w, http.StatusNotFound, "evaluation not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetEvaluation(w, r, evaluationID)
	case http.MethodPatch:
		s.handleUpdateEvaluation(w, r, evaluationID)
	case http.MethodDelete:
		s.handleDeleteEvaluation(w, r, evaluationID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleProvider(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Providers == nil {
		writeError(w, http.StatusServiceUnavailable, "provider store is not configured")
		return
	}
	providerID := strings.TrimPrefix(r.URL.Path, "/api/v1/providers/")
	providerID = strings.TrimSpace(providerID)

	if providerID == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListProviders(w, r)
		case http.MethodPost:
			s.handleCreateProvider(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(providerID, "/") {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetProvider(w, r, providerID)
	case http.MethodPatch:
		s.handleUpdateProvider(w, r, providerID)
	case http.MethodDelete:
		s.handleDeleteProvider(w, r, providerID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleRun(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Runs == nil {
		writeError(w, http.StatusServiceUnavailable, "run store is not configured")
		return
	}
	runID := strings.TrimPrefix(r.URL.Path, "/api/v1/runs/")
	runID = strings.TrimSpace(runID)

	if runID == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListRuns(w, r)
		case http.MethodPost:
			s.handleCreateRun(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(runID, "/") {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetRun(w, r, runID)
	case http.MethodPatch:
		s.handleUpdateRun(w, r, runID)
	case http.MethodDelete:
		s.handleDeleteRun(w, r, runID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, err := s.Stores.Agents.GetAgent(r.Context(), agentID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read agent")
		return
	}
	writeJSON(w, http.StatusOK, agentResponseFromRecord(*agent))
}

func (s Server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")
	var records []AgentRecord
	var total int
	var err error
	switch {
	case workspaceID != "":
		records, total, err = s.Stores.Agents.ListAgentsByWorkspace(r.Context(), workspaceID, page, limit)
	case tenantID != "":
		records, total, err = s.Stores.Agents.ListAgentsByTenant(r.Context(), tenantID, page, limit)
	default:
		records, total, err = s.Stores.Agents.ListAgents(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}
	agents := make([]AgentResponse, 0, len(records))
	for _, rec := range records {
		agents = append(agents, agentResponseFromRecord(rec))
	}
	writeJSON(w, http.StatusOK, PaginatedAgentsResponse{
		Agents: agents,
		Page:   page,
		Limit:  limit,
		Total:  total,
	})
}

func (s Server) handleGetEvaluation(w http.ResponseWriter, r *http.Request, evaluationID string) {
	evaluation, err := s.Stores.Evaluations.GetEvaluation(r.Context(), evaluationID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "evaluation not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read evaluation")
		return
	}
	writeJSON(w, http.StatusOK, evaluationResponseFromRecord(*evaluation))
}

func (s Server) handleListEvaluations(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")
	agentID := r.URL.Query().Get("agentId")
	var records []EvaluationRecord
	var total int
	var err error
	switch {
	case agentID != "":
		records, total, err = s.Stores.Evaluations.ListEvaluationsByAgent(r.Context(), agentID, page, limit)
	case workspaceID != "":
		records, total, err = s.Stores.Evaluations.ListEvaluationsByWorkspace(r.Context(), workspaceID, page, limit)
	case tenantID != "":
		records, total, err = s.Stores.Evaluations.ListEvaluationsByTenant(r.Context(), tenantID, page, limit)
	default:
		records, total, err = s.Stores.Evaluations.ListEvaluations(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list evaluations")
		return
	}
	evaluations := make([]EvaluationResponse, 0, len(records))
	for _, rec := range records {
		evaluations = append(evaluations, evaluationResponseFromRecord(rec))
	}
	writeJSON(w, http.StatusOK, PaginatedEvaluationsResponse{
		Evaluations: evaluations,
		Page:        page,
		Limit:       limit,
		Total:       total,
	})
}

func (s Server) handleGetProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := s.Stores.Providers.GetProvider(r.Context(), providerID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read provider")
		return
	}
	writeJSON(w, http.StatusOK, providerResponseFromRecord(*provider))
}

func (s Server) handleListProviders(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")
	var records []ProviderRecord
	var total int
	var err error
	switch {
	case workspaceID != "":
		records, total, err = s.Stores.Providers.ListProvidersByWorkspace(r.Context(), workspaceID, page, limit)
	case tenantID != "":
		records, total, err = s.Stores.Providers.ListProvidersByTenant(r.Context(), tenantID, page, limit)
	default:
		records, total, err = s.Stores.Providers.ListProviders(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list providers")
		return
	}
	providers := make([]ProviderResponse, 0, len(records))
	for _, rec := range records {
		providers = append(providers, providerResponseFromRecord(rec))
	}
	writeJSON(w, http.StatusOK, PaginatedProvidersResponse{
		Providers: providers,
		Page:      page,
		Limit:     limit,
		Total:     total,
	})
}

func (s Server) handleGetRun(w http.ResponseWriter, r *http.Request, runID string) {
	run, err := s.Stores.Runs.GetRun(r.Context(), runID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read run")
		return
	}
	writeJSON(w, http.StatusOK, runResponseFromRecord(*run))
}

func (s Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")
	agentID := r.URL.Query().Get("agentId")
	evaluationID := r.URL.Query().Get("evaluationId")
	var records []RunRecord
	var total int
	var err error
	switch {
	case evaluationID != "":
		records, total, err = s.Stores.Runs.ListRunsByEvaluation(r.Context(), evaluationID, page, limit)
	case agentID != "":
		records, total, err = s.Stores.Runs.ListRunsByAgent(r.Context(), agentID, page, limit)
	case workspaceID != "":
		records, total, err = s.Stores.Runs.ListRunsByWorkspace(r.Context(), workspaceID, page, limit)
	case tenantID != "":
		records, total, err = s.Stores.Runs.ListRunsByTenant(r.Context(), tenantID, page, limit)
	default:
		records, total, err = s.Stores.Runs.ListRuns(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list runs")
		return
	}
	runs := make([]RunResponse, 0, len(records))
	for _, rec := range records {
		runs = append(runs, runResponseFromRecord(rec))
	}
	writeJSON(w, http.StatusOK, PaginatedRunsResponse{
		Runs:  runs,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (s Server) handleGetTenant(w http.ResponseWriter, r *http.Request, tenantID string) {
	tenant, err := s.Stores.Tenants.GetTenant(r.Context(), tenantID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read tenant")
		return
	}
	writeJSON(w, http.StatusOK, TenantResponse{
		ID:             tenant.ID,
		OrganizationID: tenant.OrganizationID,
		Slug:           tenant.Slug,
		DisplayName:    tenant.DisplayName,
		Status:         tenant.Status,
		DefaultRegion:  tenant.DefaultRegion,
	})
}

func (s Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	records, total, err := s.Stores.Tenants.ListTenants(r.Context(), page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tenants")
		return
	}
	tenants := make([]TenantResponse, 0, len(records))
	for _, rec := range records {
		tenants = append(tenants, TenantResponse{
			ID:             rec.ID,
			OrganizationID: rec.OrganizationID,
			Slug:           rec.Slug,
			DisplayName:    rec.DisplayName,
			Status:         rec.Status,
			DefaultRegion:  rec.DefaultRegion,
		})
	}
	writeJSON(w, http.StatusOK, PaginatedTenantsResponse{
		Tenants: tenants,
		Page:    page,
		Limit:   limit,
		Total:   total,
	})
}

func workspaceResponseFromRecord(rec WorkspaceRecord) WorkspaceResponse {
	return WorkspaceResponse{
		ID:                      rec.ID,
		TenantID:                rec.TenantID,
		Slug:                    rec.Slug,
		DisplayName:             rec.DisplayName,
		Description:             rec.Description,
		Status:                  rec.Status,
		KubernetesNamespace:     rec.KubernetesNamespace,
		KubernetesWorkspaceName: rec.KubernetesWorkspaceName,
	}
}

func agentResponseFromRecord(rec AgentRecord) AgentResponse {
	return AgentResponse{
		ID:             rec.ID,
		TenantID:       rec.TenantID,
		WorkspaceID:    rec.WorkspaceID,
		Slug:           rec.Slug,
		DisplayName:    rec.DisplayName,
		Description:    rec.Description,
		Status:         rec.Status,
		Pattern:        rec.Pattern,
		RuntimeEngine:  rec.RuntimeEngine,
		RunnerClass:    rec.RunnerClass,
		ModelProvider:  rec.ModelProvider,
		ModelName:      rec.ModelName,
		LatestRevision: rec.LatestRevision,
		CompileStatus:  rec.CompileStatus,
		CompileErrors:  rec.CompileErrors,
		Revisions:      rec.Revisions,
		Spec:           rec.Spec,
	}
}

func evaluationResponseFromRecord(rec EvaluationRecord) EvaluationResponse {
	return EvaluationResponse{
		ID:               rec.ID,
		TenantID:         rec.TenantID,
		WorkspaceID:      rec.WorkspaceID,
		AgentID:          rec.AgentID,
		Slug:             rec.Slug,
		DisplayName:      rec.DisplayName,
		Description:      rec.Description,
		Status:           rec.Status,
		DatasetName:      rec.DatasetName,
		DatasetRevision:  rec.DatasetRevision,
		BaselineRevision: rec.BaselineRevision,
		Score:            rec.Score,
		GatePassed:       rec.GatePassed,
		SamplesTotal:     rec.SamplesTotal,
		SamplesEvaluated: rec.SamplesEvaluated,
		LatestRunID:      rec.LatestRunID,
		ReportRef:        rec.ReportRef,
	}
}

func providerResponseFromRecord(rec ProviderRecord) ProviderResponse {
	return ProviderResponse{
		ID:                  rec.ID,
		TenantID:            rec.TenantID,
		WorkspaceID:         rec.WorkspaceID,
		Provider:            rec.Provider,
		DisplayName:         rec.DisplayName,
		Family:              rec.Family,
		BaseURL:             rec.BaseURL,
		CredentialRef:       rec.CredentialRef,
		Status:              rec.Status,
		Domestic:            rec.Domestic,
		SupportsJSONSchema:  rec.SupportsJSONSchema,
		SupportsToolCalling: rec.SupportsToolCalling,
	}
}

func runResponseFromRecord(rec RunRecord) RunResponse {
	return RunResponse{
		ID:            rec.ID,
		TenantID:      rec.TenantID,
		WorkspaceID:   rec.WorkspaceID,
		AgentID:       rec.AgentID,
		EvaluationID:  rec.EvaluationID,
		AgentRevision: rec.AgentRevision,
		Status:        rec.Status,
		RuntimeEngine: rec.RuntimeEngine,
		RunnerClass:   rec.RunnerClass,
		StartedAt:     rec.StartedAt,
		CompletedAt:   rec.CompletedAt,
		Summary:       rec.Summary,
		TraceRef:      rec.TraceRef,
	}
}

func tenantResponseFromRecord(rec TenantRecord) TenantResponse {
	return TenantResponse{
		ID:             rec.ID,
		OrganizationID: rec.OrganizationID,
		Slug:           rec.Slug,
		DisplayName:    rec.DisplayName,
		Status:         rec.Status,
		DefaultRegion:  rec.DefaultRegion,
	}
}

func (s Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.OrganizationID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, organizationId, slug, and displayName are required")
		return
	}
	record := TenantRecord{
		ID:             req.ID,
		OrganizationID: req.OrganizationID,
		Slug:           req.Slug,
		DisplayName:    req.DisplayName,
		Status:         req.Status,
		DefaultRegion:  req.DefaultRegion,
	}
	if record.Status == "" {
		record.Status = "active"
	}
	if err := s.Stores.Tenants.CreateTenant(r.Context(), record); err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "tenant already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create tenant")
		return
	}
	if err := s.syncer().SyncTenant(r.Context(), record); err != nil {
		log.Printf("syncer: failed to sync tenant %s: %v", record.ID, err)
	}
	writeJSON(w, http.StatusCreated, tenantResponseFromRecord(record))
}

func (s Server) handleUpdateTenant(w http.ResponseWriter, r *http.Request, tenantID string) {
	var req UpdateTenantRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.DefaultRegion != nil {
		fields["default_region"] = *req.DefaultRegion
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "at least one updatable field must be provided")
		return
	}
	updated, err := s.Stores.Tenants.UpdateTenant(r.Context(), tenantID, fields)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "tenant not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tenant")
		return
	}
	if err := s.syncer().SyncTenant(r.Context(), *updated); err != nil {
		log.Printf("syncer: failed to sync tenant %s: %v", updated.ID, err)
	}
	writeJSON(w, http.StatusOK, tenantResponseFromRecord(*updated))
}

func (s Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request, tenantID string) {
	if err := s.Stores.Tenants.DeleteTenant(r.Context(), tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete tenant")
		return
	}
	if err := s.syncer().DeleteTenant(r.Context(), tenantID); err != nil {
		log.Printf("syncer: failed to delete tenant %s: %v", tenantID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req CreateAgentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	record := AgentRecord{
		ID:            req.ID,
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		Slug:          req.Slug,
		DisplayName:   req.DisplayName,
		Description:   req.Description,
		Status:        req.Status,
		Pattern:       req.Pattern,
		RuntimeEngine: req.RuntimeEngine,
		RunnerClass:   req.RunnerClass,
		ModelProvider: req.ModelProvider,
		ModelName:     req.ModelName,
		Spec:          req.Spec,
	}
	if record.Status == "" {
		record.Status = "draft"
	}
	if record.Pattern == "" {
		record.Pattern = "react"
	}
	if record.RuntimeEngine == "" {
		record.RuntimeEngine = "eino"
	}
	if record.RunnerClass == "" {
		record.RunnerClass = "adk"
	}
	if err := s.Stores.Agents.CreateAgent(r.Context(), record); err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "agent already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create agent")
		return
	}
	if err := s.syncer().SyncAgent(r.Context(), record); err != nil {
		log.Printf("syncer: failed to sync agent %s: %v", record.ID, err)
	}
	writeJSON(w, http.StatusCreated, agentResponseFromRecord(record))
}

func (s Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	var req UpdateAgentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Pattern != nil {
		fields["pattern"] = *req.Pattern
	}
	if req.RuntimeEngine != nil {
		fields["runtime_engine"] = *req.RuntimeEngine
	}
	if req.RunnerClass != nil {
		fields["runner_class"] = *req.RunnerClass
	}
	if req.ModelProvider != nil {
		fields["model_provider"] = *req.ModelProvider
	}
	if req.ModelName != nil {
		fields["model_name"] = *req.ModelName
	}
	if len(fields) == 0 && req.Spec == nil {
		writeError(w, http.StatusBadRequest, "at least one updatable field must be provided")
		return
	}
	updated, err := s.Stores.Agents.UpdateAgent(r.Context(), agentID, fields, req.Spec)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update agent")
		return
	}
	if err := s.syncer().SyncAgent(r.Context(), *updated); err != nil {
		log.Printf("syncer: failed to sync agent %s: %v", updated.ID, err)
	}
	writeJSON(w, http.StatusOK, agentResponseFromRecord(*updated))
}

func (s Server) handlePublishAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	if s.Stores.Agents == nil {
		writeError(w, http.StatusServiceUnavailable, "agent store is not configured")
		return
	}

	agent, err := s.Stores.Agents.GetAgent(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	if agent.Spec == nil {
		// No spec — mark as published with no revision.
		fields := map[string]string{
			"status":         "published",
			"compile_status": "ok",
		}
		entry := RevisionEntry{Revision: "", CreatedAt: now, Status: "ok"}
		updated, err := s.Stores.Agents.UpdateAgentPublish(r.Context(), agentID, fields, entry)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to publish agent")
			return
		}
		writeJSON(w, http.StatusOK, CompileResult{OK: true})
		_ = s.syncer().SyncAgent(r.Context(), *updated)
		return
	}

	// Compile the agent.
	k8sAgent := agentRecordToK8s(*agent)
	refs := buildReferenceIndex(agent.Spec)
	result, compileErr := compiler.CompileAgent(k8sAgent, refs)

	if compileErr != nil {
		errs := splitErrors(compileErr.Error())
		fields := map[string]string{
			"compile_status": "error",
		}
		entry := RevisionEntry{Revision: "", CreatedAt: now, Status: "error"}
		updated, err := s.Stores.Agents.UpdateAgentPublish(r.Context(), agentID, fields, entry)
		if err != nil {
			log.Printf("publish: failed to update agent %s: %v", agentID, err)
		} else if err := s.syncer().SyncAgent(r.Context(), *updated); err != nil {
			log.Printf("syncer: failed to sync agent %s: %v", agentID, err)
		}
		writeJSON(w, http.StatusOK, CompileResult{OK: false, Errors: errs})
		return
	}

	// Success — set latest revision and mark published.
	fields := map[string]string{
		"status":          "published",
		"compile_status":  "ok",
		"latest_revision": result.Revision,
	}
	entry := RevisionEntry{Revision: result.Revision, CreatedAt: now, Status: "ok"}
	updated, err := s.Stores.Agents.UpdateAgentPublish(r.Context(), agentID, fields, entry)
	if err != nil {
		log.Printf("publish: failed to update agent %s: %v", agentID, err)
	} else if err := s.syncer().SyncAgent(r.Context(), *updated); err != nil {
		log.Printf("syncer: failed to sync agent %s: %v", agentID, err)
	}

	var artifact any
	if raw, err := json.Marshal(result.Artifact); err == nil {
		_ = json.Unmarshal(raw, &artifact)
	}
	writeJSON(w, http.StatusOK, CompileResult{
		OK:       true,
		Revision: result.Revision,
		Artifact: artifact,
	})
}

func (s Server) handleDeleteAgent(w http.ResponseWriter, r *http.Request, agentID string) {
	agent, err := s.Stores.Agents.GetAgent(r.Context(), agentID)
	if errors.Is(err, ErrNotFound) {
		if err := s.Stores.Agents.DeleteAgent(r.Context(), agentID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete agent")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read agent")
		return
	}
	if err := s.Stores.Agents.DeleteAgent(r.Context(), agentID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete agent")
		return
	}
	if err := s.syncer().DeleteAgent(r.Context(), *agent); err != nil {
		log.Printf("syncer: failed to delete agent %s: %v", agentID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleCreateEvaluation(w http.ResponseWriter, r *http.Request) {
	var req CreateEvaluationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.AgentID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, agentId, slug, and displayName are required")
		return
	}
	record := EvaluationRecord{
		ID:               req.ID,
		TenantID:         req.TenantID,
		WorkspaceID:      req.WorkspaceID,
		AgentID:          req.AgentID,
		Slug:             req.Slug,
		DisplayName:      req.DisplayName,
		Description:      req.Description,
		Status:           req.Status,
		DatasetName:      req.DatasetName,
		DatasetRevision:  req.DatasetRevision,
		BaselineRevision: req.BaselineRevision,
	}
	if record.Status == "" {
		record.Status = "pending"
	}
	if err := s.Stores.Evaluations.CreateEvaluation(r.Context(), record); err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "evaluation already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create evaluation")
		return
	}
	if err := s.syncer().SyncEvaluation(r.Context(), record); err != nil {
		log.Printf("syncer: failed to sync evaluation %s: %v", record.ID, err)
	}
	writeJSON(w, http.StatusCreated, evaluationResponseFromRecord(record))
}

func (s Server) handleUpdateEvaluation(w http.ResponseWriter, r *http.Request, evaluationID string) {
	var req UpdateEvaluationRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.DatasetName != nil {
		fields["dataset_name"] = *req.DatasetName
	}
	if req.DatasetRevision != nil {
		fields["dataset_revision"] = *req.DatasetRevision
	}
	if req.BaselineRevision != nil {
		fields["baseline_revision"] = *req.BaselineRevision
	}
	if req.Score != nil {
		fields["score"] = strconv.FormatFloat(*req.Score, 'f', -1, 64)
	}
	if req.GatePassed != nil {
		fields["gate_passed"] = strconv.FormatBool(*req.GatePassed)
	}
	if req.SamplesTotal != nil {
		fields["samples_total"] = strconv.Itoa(*req.SamplesTotal)
	}
	if req.SamplesEvaluated != nil {
		fields["samples_evaluated"] = strconv.Itoa(*req.SamplesEvaluated)
	}
	if req.LatestRunID != nil {
		fields["latest_run_id"] = *req.LatestRunID
	}
	if req.ReportRef != nil {
		fields["report_ref"] = *req.ReportRef
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "at least one updatable field must be provided")
		return
	}
	updated, err := s.Stores.Evaluations.UpdateEvaluation(r.Context(), evaluationID, fields)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "evaluation not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update evaluation")
		return
	}
	if err := s.syncer().SyncEvaluation(r.Context(), *updated); err != nil {
		log.Printf("syncer: failed to sync evaluation %s: %v", updated.ID, err)
	}
	writeJSON(w, http.StatusOK, evaluationResponseFromRecord(*updated))
}

func (s Server) handleDeleteEvaluation(w http.ResponseWriter, r *http.Request, evaluationID string) {
	evaluation, err := s.Stores.Evaluations.GetEvaluation(r.Context(), evaluationID)
	if errors.Is(err, ErrNotFound) {
		if err := s.Stores.Evaluations.DeleteEvaluation(r.Context(), evaluationID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete evaluation")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read evaluation")
		return
	}
	if err := s.Stores.Evaluations.DeleteEvaluation(r.Context(), evaluationID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete evaluation")
		return
	}
	if err := s.syncer().DeleteEvaluation(r.Context(), *evaluation); err != nil {
		log.Printf("syncer: failed to delete evaluation %s: %v", evaluationID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleCreateProvider(w http.ResponseWriter, r *http.Request) {
	var req CreateProviderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.Provider == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, provider, and displayName are required")
		return
	}
	record := ProviderRecord{
		ID:                  req.ID,
		TenantID:            req.TenantID,
		WorkspaceID:         req.WorkspaceID,
		Provider:            req.Provider,
		DisplayName:         req.DisplayName,
		Family:              req.Family,
		BaseURL:             req.BaseURL,
		CredentialRef:       req.CredentialRef,
		Status:              req.Status,
		Domestic:            req.Domestic,
		SupportsJSONSchema:  req.SupportsJSONSchema,
		SupportsToolCalling: req.SupportsToolCalling,
	}
	if record.Status == "" {
		record.Status = "active"
	}
	if err := s.Stores.Providers.CreateProvider(r.Context(), record); err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "provider already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create provider")
		return
	}
	if err := s.syncer().SyncProvider(r.Context(), record); err != nil {
		log.Printf("syncer: failed to sync provider %s: %v", record.ID, err)
	}
	writeJSON(w, http.StatusCreated, providerResponseFromRecord(record))
}

func (s Server) handleUpdateProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	var req UpdateProviderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Family != nil {
		fields["family"] = *req.Family
	}
	if req.BaseURL != nil {
		fields["base_url"] = *req.BaseURL
	}
	if req.CredentialRef != nil {
		fields["credential_ref"] = *req.CredentialRef
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Domestic != nil {
		fields["domestic"] = strconv.FormatBool(*req.Domestic)
	}
	if req.SupportsJSONSchema != nil {
		fields["supports_json_schema"] = strconv.FormatBool(*req.SupportsJSONSchema)
	}
	if req.SupportsToolCalling != nil {
		fields["supports_tool_calling"] = strconv.FormatBool(*req.SupportsToolCalling)
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "at least one updatable field must be provided")
		return
	}
	updated, err := s.Stores.Providers.UpdateProvider(r.Context(), providerID, fields)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update provider")
		return
	}
	if err := s.syncer().SyncProvider(r.Context(), *updated); err != nil {
		log.Printf("syncer: failed to sync provider %s: %v", updated.ID, err)
	}
	writeJSON(w, http.StatusOK, providerResponseFromRecord(*updated))
}

func (s Server) handleDeleteProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	provider, err := s.Stores.Providers.GetProvider(r.Context(), providerID)
	if errors.Is(err, ErrNotFound) {
		if err := s.Stores.Providers.DeleteProvider(r.Context(), providerID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete provider")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read provider")
		return
	}
	if err := s.Stores.Providers.DeleteProvider(r.Context(), providerID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete provider")
		return
	}
	if err := s.syncer().DeleteProvider(r.Context(), *provider); err != nil {
		log.Printf("syncer: failed to delete provider %s: %v", providerID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s Server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req CreateRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.AgentID == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, and agentId are required")
		return
	}
	record := RunRecord{
		ID:            req.ID,
		TenantID:      req.TenantID,
		WorkspaceID:   req.WorkspaceID,
		AgentID:       req.AgentID,
		EvaluationID:  req.EvaluationID,
		AgentRevision: req.AgentRevision,
		Status:        req.Status,
		RuntimeEngine: req.RuntimeEngine,
		RunnerClass:   req.RunnerClass,
	}
	if record.Status == "" {
		record.Status = "pending"
	}
	if err := s.Stores.Runs.CreateRun(r.Context(), record); err != nil {
		if errors.Is(err, ErrConflict) {
			writeError(w, http.StatusConflict, "run already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create run")
		return
	}
	writeJSON(w, http.StatusCreated, runResponseFromRecord(record))
}

func (s Server) handleUpdateRun(w http.ResponseWriter, r *http.Request, runID string) {
	var req UpdateRunRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.StartedAt != nil {
		fields["started_at"] = *req.StartedAt
	}
	if req.CompletedAt != nil {
		fields["completed_at"] = *req.CompletedAt
	}
	if req.Summary != nil {
		fields["summary"] = *req.Summary
	}
	if req.TraceRef != nil {
		fields["trace_ref"] = *req.TraceRef
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "at least one updatable field must be provided")
		return
	}
	updated, err := s.Stores.Runs.UpdateRun(r.Context(), runID, fields)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update run")
		return
	}
	writeJSON(w, http.StatusOK, runResponseFromRecord(*updated))
}

func (s Server) handleDeleteRun(w http.ResponseWriter, r *http.Request, runID string) {
	if err := s.Stores.Runs.DeleteRun(r.Context(), runID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete run")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func paginationFromQuery(r *http.Request) (page, limit int) {
	page, limit = 1, 20
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	return
}

// PromptTemplate handlers

func (s Server) handlePromptTemplate(w http.ResponseWriter, r *http.Request) {
	if s.Stores.PromptTemplates == nil {
		writeError(w, http.StatusServiceUnavailable, "prompt template store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/prompt-templates/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListPromptTemplates(w, r)
		case http.MethodPost:
			s.handleCreatePromptTemplate(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "prompt template not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetPromptTemplate(w, r, id)
	case http.MethodPatch:
		s.handleUpdatePromptTemplate(w, r, id)
	case http.MethodDelete:
		s.handleDeletePromptTemplate(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetPromptTemplate(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.PromptTemplates.GetPromptTemplate(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "prompt template not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, PromptTemplateResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Template: rec.Template, Variables: rec.Variables,
	})
}

func (s Server) handleListPromptTemplates(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []PromptTemplateRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.PromptTemplates.ListPromptTemplatesByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.PromptTemplates.ListPromptTemplatesByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.PromptTemplates.ListPromptTemplates(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]PromptTemplateResponse, len(records))
	for i, rec := range records {
		resp[i] = PromptTemplateResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, Template: rec.Template, Variables: rec.Variables,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedPromptTemplatesResponse{Templates: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreatePromptTemplate(w http.ResponseWriter, r *http.Request) {
	var req CreatePromptTemplateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := PromptTemplateRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, Template: req.Template, Variables: req.Variables,
	}
	if rec.Status == "" {
		rec.Status = "draft"
	}
	if err := s.Stores.PromptTemplates.CreatePromptTemplate(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "prompt template already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, PromptTemplateResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Template: rec.Template, Variables: rec.Variables,
	})
}

func (s Server) handleUpdatePromptTemplate(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdatePromptTemplateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Template != nil {
		fields["template"] = *req.Template
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.PromptTemplates.UpdatePromptTemplate(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "prompt template not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, PromptTemplateResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Template: rec.Template, Variables: rec.Variables,
	})
}

func (s Server) handleDeletePromptTemplate(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.PromptTemplates.DeletePromptTemplate(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ToolProvider handlers

func (s Server) handleToolProvider(w http.ResponseWriter, r *http.Request) {
	if s.Stores.ToolProviders == nil {
		writeError(w, http.StatusServiceUnavailable, "tool provider store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/tool-providers/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListToolProviders(w, r)
		case http.MethodPost:
			s.handleCreateToolProvider(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "tool provider not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetToolProvider(w, r, id)
	case http.MethodPatch:
		s.handleUpdateToolProvider(w, r, id)
	case http.MethodDelete:
		s.handleDeleteToolProvider(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetToolProvider(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.ToolProviders.GetToolProvider(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "tool provider not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolProviderResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, ToolType: rec.ToolType, Endpoint: rec.Endpoint, Config: rec.Config,
	})
}

func (s Server) handleListToolProviders(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []ToolProviderRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.ToolProviders.ListToolProvidersByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.ToolProviders.ListToolProvidersByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.ToolProviders.ListToolProviders(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]ToolProviderResponse, len(records))
	for i, rec := range records {
		resp[i] = ToolProviderResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, ToolType: rec.ToolType, Endpoint: rec.Endpoint, Config: rec.Config,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedToolProvidersResponse{Tools: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreateToolProvider(w http.ResponseWriter, r *http.Request) {
	var req CreateToolProviderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := ToolProviderRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, ToolType: req.ToolType, Endpoint: req.Endpoint, Config: req.Config,
	}
	if rec.Status == "" {
		rec.Status = "active"
	}
	if err := s.Stores.ToolProviders.CreateToolProvider(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "tool provider already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ToolProviderResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, ToolType: rec.ToolType, Endpoint: rec.Endpoint, Config: rec.Config,
	})
}

func (s Server) handleUpdateToolProvider(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateToolProviderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.ToolType != nil {
		fields["tool_type"] = *req.ToolType
	}
	if req.Endpoint != nil {
		fields["endpoint"] = *req.Endpoint
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.ToolProviders.UpdateToolProvider(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "tool provider not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ToolProviderResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, ToolType: rec.ToolType, Endpoint: rec.Endpoint, Config: rec.Config,
	})
}

func (s Server) handleDeleteToolProvider(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.ToolProviders.DeleteToolProvider(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// KnowledgeBase handlers

func (s Server) handleKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	if s.Stores.KnowledgeBases == nil {
		writeError(w, http.StatusServiceUnavailable, "knowledge base store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge-bases/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListKnowledgeBases(w, r)
		case http.MethodPost:
			s.handleCreateKnowledgeBase(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "knowledge base not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetKnowledgeBase(w, r, id)
	case http.MethodPatch:
		s.handleUpdateKnowledgeBase(w, r, id)
	case http.MethodDelete:
		s.handleDeleteKnowledgeBase(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetKnowledgeBase(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.KnowledgeBases.GetKnowledgeBase(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "knowledge base not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, KnowledgeBaseResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, SourceType: rec.SourceType, SourceRef: rec.SourceRef,
		EmbedModel: rec.EmbedModel, ChunkSize: rec.ChunkSize, ChunkOverlap: rec.ChunkOverlap,
	})
}

func (s Server) handleListKnowledgeBases(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []KnowledgeBaseRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.KnowledgeBases.ListKnowledgeBasesByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.KnowledgeBases.ListKnowledgeBasesByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.KnowledgeBases.ListKnowledgeBases(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]KnowledgeBaseResponse, len(records))
	for i, rec := range records {
		resp[i] = KnowledgeBaseResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, SourceType: rec.SourceType, SourceRef: rec.SourceRef,
			EmbedModel: rec.EmbedModel, ChunkSize: rec.ChunkSize, ChunkOverlap: rec.ChunkOverlap,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedKnowledgeBasesResponse{Bases: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreateKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	var req CreateKnowledgeBaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := KnowledgeBaseRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, SourceType: req.SourceType, SourceRef: req.SourceRef,
		EmbedModel: req.EmbedModel, ChunkSize: req.ChunkSize, ChunkOverlap: req.ChunkOverlap,
	}
	if rec.Status == "" {
		rec.Status = "active"
	}
	if err := s.Stores.KnowledgeBases.CreateKnowledgeBase(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "knowledge base already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, KnowledgeBaseResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, SourceType: rec.SourceType, SourceRef: rec.SourceRef,
		EmbedModel: rec.EmbedModel, ChunkSize: rec.ChunkSize, ChunkOverlap: rec.ChunkOverlap,
	})
}

func (s Server) handleUpdateKnowledgeBase(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateKnowledgeBaseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.SourceType != nil {
		fields["source_type"] = *req.SourceType
	}
	if req.SourceRef != nil {
		fields["source_ref"] = *req.SourceRef
	}
	if req.EmbedModel != nil {
		fields["embed_model"] = *req.EmbedModel
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.KnowledgeBases.UpdateKnowledgeBase(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "knowledge base not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, KnowledgeBaseResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, SourceType: rec.SourceType, SourceRef: rec.SourceRef,
		EmbedModel: rec.EmbedModel, ChunkSize: rec.ChunkSize, ChunkOverlap: rec.ChunkOverlap,
	})
}

func (s Server) handleDeleteKnowledgeBase(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.KnowledgeBases.DeleteKnowledgeBase(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Dataset handlers

func (s Server) handleDataset(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Datasets == nil {
		writeError(w, http.StatusServiceUnavailable, "dataset store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/datasets/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListDatasets(w, r)
		case http.MethodPost:
			s.handleCreateDataset(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "dataset not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetDataset(w, r, id)
	case http.MethodPatch:
		s.handleUpdateDataset(w, r, id)
	case http.MethodDelete:
		s.handleDeleteDataset(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetDataset(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.Datasets.GetDataset(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, DatasetResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Format: rec.Format, SourceRef: rec.SourceRef,
		RowCount: rec.RowCount, Columns: rec.Columns,
	})
}

func (s Server) handleListDatasets(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []DatasetRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.Datasets.ListDatasetsByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.Datasets.ListDatasetsByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.Datasets.ListDatasets(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]DatasetResponse, len(records))
	for i, rec := range records {
		resp[i] = DatasetResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, Format: rec.Format, SourceRef: rec.SourceRef,
			RowCount: rec.RowCount, Columns: rec.Columns,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedDatasetsResponse{Datasets: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreateDataset(w http.ResponseWriter, r *http.Request) {
	var req CreateDatasetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := DatasetRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, Format: req.Format, SourceRef: req.SourceRef,
		RowCount: req.RowCount, Columns: req.Columns,
	}
	if rec.Status == "" {
		rec.Status = "active"
	}
	if err := s.Stores.Datasets.CreateDataset(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "dataset already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, DatasetResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Format: rec.Format, SourceRef: rec.SourceRef,
		RowCount: rec.RowCount, Columns: rec.Columns,
	})
}

func (s Server) handleUpdateDataset(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateDatasetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Format != nil {
		fields["format"] = *req.Format
	}
	if req.SourceRef != nil {
		fields["source_ref"] = *req.SourceRef
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.Datasets.UpdateDataset(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "dataset not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, DatasetResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Format: rec.Format, SourceRef: rec.SourceRef,
		RowCount: rec.RowCount, Columns: rec.Columns,
	})
}

func (s Server) handleDeleteDataset(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.Datasets.DeleteDataset(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// MCPServer handlers

func (s Server) handleMCPServer(w http.ResponseWriter, r *http.Request) {
	if s.Stores.MCPServers == nil {
		writeError(w, http.StatusServiceUnavailable, "mcp server store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp-servers/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListMCPServers(w, r)
		case http.MethodPost:
			s.handleCreateMCPServer(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "mcp server not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetMCPServer(w, r, id)
	case http.MethodPatch:
		s.handleUpdateMCPServer(w, r, id)
	case http.MethodDelete:
		s.handleDeleteMCPServer(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetMCPServer(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.MCPServers.GetMCPServer(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "mcp server not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, MCPServerResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Endpoint: rec.Endpoint, Transport: rec.Transport,
		Version: rec.Version,
	})
}

func (s Server) handleListMCPServers(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []MCPServerRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.MCPServers.ListMCPServersByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.MCPServers.ListMCPServersByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.MCPServers.ListMCPServers(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]MCPServerResponse, len(records))
	for i, rec := range records {
		resp[i] = MCPServerResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, Endpoint: rec.Endpoint, Transport: rec.Transport,
			Version: rec.Version,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedMCPServersResponse{Servers: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreateMCPServer(w http.ResponseWriter, r *http.Request) {
	var req CreateMCPServerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := MCPServerRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, Endpoint: req.Endpoint, Transport: req.Transport,
		Version: req.Version,
	}
	if rec.Status == "" {
		rec.Status = "active"
	}
	if err := s.Stores.MCPServers.CreateMCPServer(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "mcp server already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, MCPServerResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Endpoint: rec.Endpoint, Transport: rec.Transport,
		Version: rec.Version,
	})
}

func (s Server) handleUpdateMCPServer(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateMCPServerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.Endpoint != nil {
		fields["endpoint"] = *req.Endpoint
	}
	if req.Transport != nil {
		fields["transport"] = *req.Transport
	}
	if req.Version != nil {
		fields["version"] = *req.Version
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.MCPServers.UpdateMCPServer(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "mcp server not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, MCPServerResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, Endpoint: rec.Endpoint, Transport: rec.Transport,
		Version: rec.Version,
	})
}

func (s Server) handleDeleteMCPServer(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.MCPServers.DeleteMCPServer(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AgentPolicy handlers

func (s Server) handleAgentPolicy(w http.ResponseWriter, r *http.Request) {
	if s.Stores.AgentPolicies == nil {
		writeError(w, http.StatusServiceUnavailable, "agent policy store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/agent-policies/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListAgentPolicies(w, r)
		case http.MethodPost:
			s.handleCreateAgentPolicy(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "agent policy not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetAgentPolicy(w, r, id)
	case http.MethodPatch:
		s.handleUpdateAgentPolicy(w, r, id)
	case http.MethodDelete:
		s.handleDeleteAgentPolicy(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetAgentPolicy(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.AgentPolicies.GetAgentPolicy(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "agent policy not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentPolicyResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, PolicyType: rec.PolicyType, Rules: rec.Rules,
		Enforcement: rec.Enforcement,
	})
}

func (s Server) handleListAgentPolicies(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []AgentPolicyRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.AgentPolicies.ListAgentPoliciesByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.AgentPolicies.ListAgentPoliciesByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.AgentPolicies.ListAgentPolicies(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]AgentPolicyResponse, len(records))
	for i, rec := range records {
		resp[i] = AgentPolicyResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, PolicyType: rec.PolicyType, Rules: rec.Rules,
			Enforcement: rec.Enforcement,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedAgentPoliciesResponse{Policies: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreateAgentPolicy(w http.ResponseWriter, r *http.Request) {
	var req CreateAgentPolicyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := AgentPolicyRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, PolicyType: req.PolicyType, Rules: req.Rules,
		Enforcement: req.Enforcement,
	}
	if rec.Status == "" {
		rec.Status = "active"
	}
	if err := s.Stores.AgentPolicies.CreateAgentPolicy(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "agent policy already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, AgentPolicyResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, PolicyType: rec.PolicyType, Rules: rec.Rules,
		Enforcement: rec.Enforcement,
	})
}

func (s Server) handleUpdateAgentPolicy(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateAgentPolicyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.PolicyType != nil {
		fields["policy_type"] = *req.PolicyType
	}
	if req.Enforcement != nil {
		fields["enforcement"] = *req.Enforcement
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.AgentPolicies.UpdateAgentPolicy(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "agent policy not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, AgentPolicyResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, PolicyType: rec.PolicyType, Rules: rec.Rules,
		Enforcement: rec.Enforcement,
	})
}

func (s Server) handleDeleteAgentPolicy(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.AgentPolicies.DeleteAgentPolicy(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Skill handlers

func (s Server) handleSkill(w http.ResponseWriter, r *http.Request) {
	if s.Stores.Skills == nil {
		writeError(w, http.StatusServiceUnavailable, "skill store is not configured")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	id = strings.TrimSpace(id)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListSkills(w, r)
		case http.MethodPost:
			s.handleCreateSkill(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method must be GET or POST")
		}
		return
	}

	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetSkill(w, r, id)
	case http.MethodPatch:
		s.handleUpdateSkill(w, r, id)
	case http.MethodDelete:
		s.handleDeleteSkill(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method must be GET, PATCH, or DELETE")
	}
}

func (s Server) handleGetSkill(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.Stores.Skills.GetSkill(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, SkillResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, SkillType: rec.SkillType, Entrypoint: rec.Entrypoint,
		Config: rec.Config,
	})
}

func (s Server) handleListSkills(w http.ResponseWriter, r *http.Request) {
	page, limit := paginationFromQuery(r)
	tenantID := r.URL.Query().Get("tenantId")
	workspaceID := r.URL.Query().Get("workspaceId")

	var records []SkillRecord
	var total int
	var err error

	if tenantID != "" {
		records, total, err = s.Stores.Skills.ListSkillsByTenant(r.Context(), tenantID, page, limit)
	} else if workspaceID != "" {
		records, total, err = s.Stores.Skills.ListSkillsByWorkspace(r.Context(), workspaceID, page, limit)
	} else {
		records, total, err = s.Stores.Skills.ListSkills(r.Context(), page, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := make([]SkillResponse, len(records))
	for i, rec := range records {
		resp[i] = SkillResponse{
			ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
			Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
			Status: rec.Status, SkillType: rec.SkillType, Entrypoint: rec.Entrypoint,
			Config: rec.Config,
		}
	}
	writeJSON(w, http.StatusOK, PaginatedSkillsResponse{Skills: resp, Page: page, Limit: limit, Total: total})
}

func (s Server) handleCreateSkill(w http.ResponseWriter, r *http.Request) {
	var req CreateSkillRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ID == "" || req.TenantID == "" || req.WorkspaceID == "" || req.Slug == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "id, tenantId, workspaceId, slug, and displayName are required")
		return
	}
	rec := SkillRecord{
		ID: req.ID, TenantID: req.TenantID, WorkspaceID: req.WorkspaceID,
		Slug: req.Slug, DisplayName: req.DisplayName, Description: req.Description,
		Status: req.Status, SkillType: req.SkillType, Entrypoint: req.Entrypoint,
		Config: req.Config,
	}
	if rec.Status == "" {
		rec.Status = "active"
	}
	if err := s.Stores.Skills.CreateSkill(r.Context(), rec); err != nil {
		if err == ErrConflict {
			writeError(w, http.StatusConflict, "skill already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, SkillResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, SkillType: rec.SkillType, Entrypoint: rec.Entrypoint,
		Config: rec.Config,
	})
}

func (s Server) handleUpdateSkill(w http.ResponseWriter, r *http.Request, id string) {
	var req UpdateSkillRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	fields := map[string]string{}
	if req.DisplayName != nil {
		fields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		fields["description"] = *req.Description
	}
	if req.Status != nil {
		fields["status"] = *req.Status
	}
	if req.SkillType != nil {
		fields["skill_type"] = *req.SkillType
	}
	if req.Entrypoint != nil {
		fields["entrypoint"] = *req.Entrypoint
	}
	if len(fields) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	rec, err := s.Stores.Skills.UpdateSkill(r.Context(), id, fields)
	if err != nil {
		if err == ErrNotFound {
			writeError(w, http.StatusNotFound, "skill not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, SkillResponse{
		ID: rec.ID, TenantID: rec.TenantID, WorkspaceID: rec.WorkspaceID,
		Slug: rec.Slug, DisplayName: rec.DisplayName, Description: rec.Description,
		Status: rec.Status, SkillType: rec.SkillType, Entrypoint: rec.Entrypoint,
		Config: rec.Config,
	})
}

func (s Server) handleDeleteSkill(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.Stores.Skills.DeleteSkill(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		_, _ = fmt.Fprintf(w, `{"error":"failed to encode response"}`)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
