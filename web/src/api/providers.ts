import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { PaginatedProvidersResponse, ProviderAccount, CreateProviderRequest, UpdateProviderRequest } from "@/types/api";
import { api } from "./client";

export function useProviders(page: number, limit: number, tenantId?: string, workspaceId?: string) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (tenantId) params.set("tenantId", tenantId);
  if (workspaceId) params.set("workspaceId", workspaceId);

  return useQuery({
    queryKey: ["providers", page, limit, tenantId, workspaceId],
    queryFn: () => api.get<PaginatedProvidersResponse>(`/providers/?${params.toString()}`),
  });
}

export function useProvider(id: string | undefined) {
  return useQuery({
    queryKey: ["providers", id],
    queryFn: () => api.get<ProviderAccount>(`/providers/${id}`),
    enabled: !!id,
  });
}

export function useCreateProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProviderRequest) => api.post<ProviderAccount>("/providers/", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["providers"] });
    },
  });
}

export function useUpdateProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateProviderRequest) =>
      api.patch<ProviderAccount>(`/providers/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["providers"] });
    },
  });
}

export function useDeleteProvider() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/providers/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["providers"] });
    },
  });
}
