import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../../app/auth";
import { Button } from "../ui/button";

export function ProtectedLayout() {
  const { token, user, logout } = useAuth();
  const location = useLocation();

  if (!token || !user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return (
    <div className="min-h-screen">
      <nav className="sticky top-0 z-10 flex flex-col gap-4 border-b border-white/8 bg-[rgba(19,59,51,0.92)] px-4 py-4 text-[#f9f4ec] backdrop-blur-xl sm:flex-row sm:items-center sm:justify-between sm:px-5">
        <div className="flex items-center gap-4">
          <div className="grid size-10 place-items-center rounded-2xl bg-[linear-gradient(135deg,#f4efe7,#d6bfa5)] font-extrabold text-[var(--panel)]">
            TF
          </div>
          <div>
            <strong className="block text-lg font-semibold tracking-tight">TaskFlow</strong>
            <p className="text-sm text-[#f1e6d7]/78">
              Projects, tasks, and ownership in one clean workspace.
            </p>
          </div>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <div className="rounded-full border border-white/12 bg-white/8 px-3 py-2 text-sm font-medium text-[#f6ebdc]">
            {user.name}
          </div>
          <Button
            className="h-10 rounded-full border border-white/18 bg-transparent px-4 text-[#f9f4ec] hover:bg-white/10 hover:text-[#f9f4ec]"
            variant="ghost"
            onClick={logout}
            type="button"
          >
            Logout
          </Button>
        </div>
      </nav>
      <Outlet />
    </div>
  );
}
