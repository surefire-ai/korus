import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import type { PaginatedTenantsResponse, Tenant, CreateTenantRequest, UpdateTenantRequest } from "@/types/api";
import { api } from "./client";

export function useTenants(page: number, limit: number) {
  return useQuery({
    queryKey: ["tenants", page, limit],
    queryFn: () => api.get<PaginatedTenantsResponse>(`/tenants/?page=${page}&limit=${limit}`),
  });
}

export function useTenant(id: string | undefined) {
  return useQuery({
    queryKey: ["tenants", id],
    queryFn: () => api.get<Tenant>(`/tenants/${id}`),
    enabled: !!id,
  });
}

export function useCreateTenant() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateTenantRequest) => api.post<Tenant>("/tenants/", data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tenants"] });
    },
  });
}

export function useUpdateTenant() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & UpdateTenantRequest) =>
      api.patch<Tenant>(`/tenants/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tenants"] });
    },
  });
}

export function useDeleteTenant() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api.delete<void>(`/tenants/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tenants"] });
    },
  });
}
