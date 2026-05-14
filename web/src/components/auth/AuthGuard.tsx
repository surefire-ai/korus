import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useCurrentUser } from "@/api/auth";
import { LoadingSkeleton } from "@/components/shared/LoadingSkeleton";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const navigate = useNavigate();
  const { data: user, isLoading, isError } = useCurrentUser();

  useEffect(() => {
    if (isError) {
      navigate("/login", { replace: true });
    }
  }, [isError, navigate]);

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <LoadingSkeleton />
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return <>{children}</>;
}
