import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useAuth, useApi } from "../app/auth";
import { ProjectDetailState } from "../components/project-detail/ProjectDetailState";
import { ProjectSidebar } from "../components/project-detail/ProjectSidebar";
import { TaskFilterBar } from "../components/project-detail/TaskFilterBar";
import { TaskBoard } from "../components/project-detail/TaskBoard";
import { TaskModal } from "../components/tasks/TaskModal";
import { useProjectDetailQueryController } from "../features/project-detail/controllers/useProjectDetailQueryController";
import { useProjectDetailTaskController } from "../features/project-detail/controllers/useProjectDetailTaskController";
import { useProjectDetailViewModel } from "../features/project-detail/hooks/useProjectDetailViewModel";
import { getProjectStats } from "../lib/services/projects";
import { ProjectStats, TaskStatus } from "../lib/types";

export function ProjectDetailPage() {
  const { projectId = "" } = useParams();
  const { user } = useAuth();
  const api = useApi();
  const navigate = useNavigate();
  const [stats, setStats] = useState<ProjectStats | null>(null);
  const {
    users,
    project,
    allTasks,
    tasks,
    loadingProject,
    loadingTasks,
    projectError,
    modalState,
    deletingTaskId,
    statusSavingId,
    visibleColumns,
    assigneeOptions,
    userMap,
    statusFilter,
    assigneeFilter,
    setStatusFilter,
    setAssigneeFilter,
    openCreateModal,
    openEditModal,
    closeModal
  } = useProjectDetailViewModel();
  const { refreshProjectAndTasks } = useProjectDetailQueryController(projectId);
  const { handleDeleteTask, handleStatusChange, handleTaskSaved } =
    useProjectDetailTaskController(projectId, refreshProjectAndTasks);

  useEffect(() => {
    if (!projectId) return;
    getProjectStats(api, projectId)
      .then(setStats)
      .catch(() => setStats(null));
  }, [projectId, allTasks.length]);

  if (loadingProject) {
    return <ProjectDetailState message="Loading project..." />;
  }

  if (projectError || !project) {
    return (
      <ProjectDetailState
        actionLabel="Back to projects"
        message={projectError || "Project not found."}
        onAction={() => navigate("/")}
        tone="destructive"
      />
    );
  }

  const canEditProject = project.owner_id === user?.id;

  return (
    <div className="flex min-h-[calc(100vh-60px)]">
      {/* Sticky sidebar */}
      <ProjectSidebar
        project={project}
        stats={stats}
        roleLabel={canEditProject ? "Owner" : "Member"}
        allTaskCount={allTasks.length}
        visibleTaskCount={tasks.length}
      />

      {/* Main content */}
      <div className="flex min-w-0 flex-1 flex-col">
        {/* Mobile-only project header (sidebar is hidden on mobile) */}
        <div className="flex items-center gap-3 border-b border-border bg-card px-4 py-3 md:hidden">
          <button
            className="text-sm font-medium text-primary"
            onClick={() => navigate("/")}
            type="button"
          >
            ← Back
          </button>
          <span className="truncate font-semibold">{project.name}</span>
          <span className="ml-auto shrink-0 rounded-full bg-muted px-2 py-0.5 text-xs font-medium">
            {canEditProject ? "Owner" : "Member"}
          </span>
        </div>

        <TaskFilterBar
          statusFilter={statusFilter}
          assigneeFilter={assigneeFilter}
          assigneeOptions={assigneeOptions}
          currentUserId={user?.id}
          userMap={userMap}
          taskError=""
          onStatusChange={setStatusFilter}
          onAssigneeChange={setAssigneeFilter}
          onCreateTask={openCreateModal}
          onRefresh={() => void refreshProjectAndTasks()}
        />

        <TaskBoard
          columns={visibleColumns}
          currentUserId={user?.id}
          deletingTaskId={deletingTaskId}
          loading={loadingTasks}
          onCardClick={openEditModal}
          onDeleteTask={(taskId) => void handleDeleteTask(taskId)}
          onStatusChange={(task, status) => void handleStatusChange(task, status as TaskStatus)}
          statusSavingId={statusSavingId}
          tasks={tasks}
          userMap={userMap}
        />
      </div>

      {modalState ? (
        <TaskModal
          key={modalState.mode === "edit" ? modalState.task.id : "new-task"}
          mode={modalState.mode}
          onClose={closeModal}
          onSaved={() => void handleTaskSaved()}
          onDelete={
            modalState.mode === "edit"
              ? () => void handleDeleteTask(modalState.task.id).then(closeModal)
              : undefined
          }
          projectId={projectId}
          task={modalState.task}
          users={users}
        />
      ) : null}
    </div>
  );
}
