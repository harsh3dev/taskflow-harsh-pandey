import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AppErrorBoundary } from "./AppErrorBoundary";
import { AuthProvider } from "./auth";
import { ProtectedLayout } from "../components/layout/ProtectedLayout";
import { AuthPage } from "../pages/AuthPage";
import { ProjectsPage } from "../pages/ProjectsPage";
import { ProjectDetailPage } from "../pages/ProjectDetailPage";

export default function App() {
  return (
    <AppErrorBoundary>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<AuthPage mode="login" />} />
            <Route path="/register" element={<AuthPage mode="register" />} />
            <Route element={<ProtectedLayout />}>
              <Route index element={<ProjectsPage />} />
              <Route path="/projects/:projectId" element={<ProjectDetailPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </AppErrorBoundary>
  );
}
