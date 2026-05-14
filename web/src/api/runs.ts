import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { PaginatedRunsResponse, Run, CreateRunRequest, UpdateRunRequest } from "@/types/api";
import { api } from "./client";

export function useRuns(page: number, limit: number, tenantId?: string, workspaceId?: string, agentId?: string, evaluationId?: string) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (tenantId) params.set("tenantId", tenantId);
  if (workspaceId) params.set("workspaceId", workspaceId);
  if (agentId) params.set("agentId", agentId);
  if (evaluationId) params.set("evaluationId", evaluationId);

  return useQuery({
    queryKey: ["runs", page, limit, tenantId, workspaceId, agentId, evaluationId],
    queryFn: () => api.get<PaginatedRunsResponse>(`/runs/?${params.toString()}`),
  });
}

export function useRun(id: string | undefined) {
  return useQuery({
    queryKey: ["runs", id],
    queryFn: () => api.get<Run>(`/runs/${id}`),
    enabled: !!id,
  });
}

export function useCreateRun() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateRunRequest) => api.post<Run>("/runs/", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["runs"] });
    },
  });
}

export function useUpdateRun() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateRunRequest) =>
      api.patch<Run>(`/runs/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["runs"] });
    },
  });
}

export function useDeleteRun() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/runs/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["runs"] });
    },
  });
}
