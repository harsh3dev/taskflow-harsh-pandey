import { useEffect, useState } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../../app/auth";
import { SidebarProvider, useSidebar } from "../../app/sidebar";
import { Button } from "../ui/button";
import { applyTheme, getStoredTheme, toggleTheme, type Theme } from "../../lib/theme";

function NavBar() {
  const { user, logout } = useAuth();
  const { isOpen, toggle } = useSidebar();
  const [theme, setTheme] = useState<Theme>(() => getStoredTheme());

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  function handleToggleTheme() {
    const next = toggleTheme(theme);
    setTheme(next);
    applyTheme(next);
  }

  return (
    <nav className="sticky top-0 z-10 flex items-center gap-3 border-b border-navbar-foreground/15 bg-navbar px-4 py-3 text-navbar-foreground shadow-sm sm:px-5">
      {/* Hamburger — mobile only */}
      <button
        aria-label={isOpen ? "Close menu" : "Open menu"}
        className="flex size-8 shrink-0 flex-col items-center justify-center gap-1.5 rounded-lg text-navbar-foreground hover:bg-navbar-foreground/10 md:hidden"
        onClick={toggle}
        type="button"
      >
        {isOpen ? (
          <span className="text-lg leading-none">✕</span>
        ) : (
          <>
            <span className="h-0.5 w-5 rounded-full bg-current" />
            <span className="h-0.5 w-5 rounded-full bg-current" />
            <span className="h-0.5 w-5 rounded-full bg-current" />
          </>
        )}
      </button>

      {/* Logo */}
      <div className="flex items-center gap-3">
        <div className="hidden size-9 shrink-0 place-items-center rounded-xl bg-navbar-foreground/20 font-extrabold text-navbar-foreground md:grid">
          TF
        </div>
        <div className="hidden sm:block">
          <strong className="block text-base font-semibold tracking-tight">TaskFlow</strong>
          <p className="text-xs text-navbar-foreground/70">
            Projects, tasks, and ownership in one clean workspace.
          </p>
        </div>
        <strong className="block text-base font-semibold tracking-tight sm:hidden">TaskFlow</strong>
      </div>

      {/* Right actions */}
      <div className="ml-auto flex items-center gap-2">
        <div className="rounded-full border border-navbar-foreground/20 bg-navbar-foreground/10 px-3 py-1.5 text-sm font-medium text-navbar-foreground">
          {user?.name}
        </div>
        <Button
          aria-label={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
          className="size-9 rounded-full border border-navbar-foreground/20 bg-transparent p-0 text-navbar-foreground hover:bg-navbar-foreground/10 hover:text-navbar-foreground"
          variant="ghost"
          onClick={handleToggleTheme}
          type="button"
          title={theme === "dark" ? "Light mode" : "Dark mode"}
        >
          {theme === "dark" ? "☀️" : "🌙"}
        </Button>
        <Button
          className="hidden h-9 rounded-full border border-navbar-foreground/20 bg-transparent px-4 text-navbar-foreground hover:bg-navbar-foreground/10 hover:text-navbar-foreground sm:flex"
          variant="ghost"
          onClick={logout}
          type="button"
        >
          Logout
        </Button>
      </div>
    </nav>
  );
}

function Layout() {
  const { token, user } = useAuth();
  const location = useLocation();
  const { isOpen, close } = useSidebar();

  if (!token || !user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-background to-muted/30">
      <NavBar />
      <div className="relative">
        {/* Mobile backdrop */}
        {isOpen && (
          <div
            className="fixed inset-0 z-40 bg-black/40 md:hidden"
            onClick={close}
            aria-hidden="true"
          />
        )}
        <Outlet />
      </div>
    </div>
  );
}

export function ProtectedLayout() {
  return (
    <SidebarProvider>
      <Layout />
    </SidebarProvider>
  );
}
