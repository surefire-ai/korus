import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";
import type { User } from "@/types/api";

export function useCurrentUser() {
  return useQuery<User>({
    queryKey: ["auth", "me"],
    queryFn: () => api.get<User>("/auth/me"),
    retry: false,
    staleTime: 5 * 60 * 1000,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { username: string; password: string }) =>
      api.post<User>("/auth/login", data),
    onSuccess: (user) => {
      queryClient.setQueryData(["auth", "me"], user);
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => api.post<void>("/auth/logout", {}),
    onSuccess: () => {
      queryClient.setQueryData(["auth", "me"], null);
      queryClient.clear();
    },
  });
}
