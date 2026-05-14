package manager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompileAgentReturnsOKForValidSpec(t *testing.T) {
	handler := Server{
		Stores: Stores{
			Agents: &fakeAgentStore{
				records: map[string]AgentRecord{
					"agent_1": {
						ID: "agent_1", TenantID: "t_1", WorkspaceID: "ws_1", Slug: "test",
						DisplayName: "Test Agent", Status: "draft", Pattern: "react",
						RuntimeEngine: "eino", RunnerClass: "adk",
						Spec: &AgentSpecData{
							Models: map[string]ModelConfig{
								"planner": {Provider: "openai", Model: "gpt-4o"},
							},
							Pattern: &PatternConfig{Type: "react", ModelRef: "planner"},
						},
					},
				},
				orderedIDs: []string{"agent_1"},
			},
		},
	}.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent_1/compile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result CompileResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected ok=true, got errors: %v", result.Errors)
	}
	if result.Revision == "" {
		t.Fatal("expected non-empty revision")
	}
}

func TestCompileAgentReturnsErrorsForInvalidWorkflow(t *testing.T) {
	handler := Server{
		Stores: Stores{
			Agents: &fakeAgentStore{
				records: map[string]AgentRecord{
					"agent_1": {
						ID: "agent_1", TenantID: "t_1", WorkspaceID: "ws_1", Slug: "test",
						DisplayName: "Test Agent", Status: "draft", Pattern: "workflow",
						RuntimeEngine: "eino", RunnerClass: "adk",
						Spec: &AgentSpecData{
							Models: map[string]ModelConfig{
								"planner": {Provider: "openai", Model: "gpt-4o"},
							},
							Pattern: &PatternConfig{Type: "workflow"},
							Graph: &GraphConfig{
								Nodes: []GraphNode{
									{Name: "start", Kind: "start"},
									{Name: "end", Kind: "end"},
								},
								Edges: []GraphEdge{
									{From: "start", To: "end"},
								},
							},
						},
					},
				},
				orderedIDs: []string{"agent_1"},
			},
		},
	}.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent_1/compile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var result CompileResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.OK {
		t.Fatal("expected ok=false for workflow with only terminal nodes")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestCompileAgentReturns404ForMissingAgent(t *testing.T) {
	handler := Server{
		Stores: Stores{
			Agents: &fakeAgentStore{
				records:    map[string]AgentRecord{},
				orderedIDs: []string{},
			},
		},
	}.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/no_such/compile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCompileAgentRejectsGetMethod(t *testing.T) {
	handler := Server{
		Stores: Stores{
			Agents: &fakeAgentStore{
				records:    map[string]AgentRecord{},
				orderedIDs: []string{},
			},
		},
	}.Handler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agents/agent_1/compile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestCompileAgentRequiresStore(t *testing.T) {
	handler := Server{Stores: Stores{}}.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/agent_1/compile", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
