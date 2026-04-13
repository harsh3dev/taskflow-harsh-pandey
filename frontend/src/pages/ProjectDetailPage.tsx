import { useNavigate, useParams } from "react-router-dom";
import { useAuth } from "../app/auth";
import { ProjectDetailFiltersCard } from "../components/project-detail/ProjectDetailFiltersCard";
import { ProjectDetailHeader } from "../components/project-detail/ProjectDetailHeader";
import { ProjectDetailState } from "../components/project-detail/ProjectDetailState";
import { TaskBoard } from "../components/project-detail/TaskBoard";
import { TaskModal } from "../components/tasks/TaskModal";
import { useProjectDetailQueryController } from "../features/project-detail/controllers/useProjectDetailQueryController";
import { useProjectDetailTaskController } from "../features/project-detail/controllers/useProjectDetailTaskController";
import { useProjectDetailViewModel } from "../features/project-detail/hooks/useProjectDetailViewModel";
import { TaskStatus } from "../lib/types";

export function ProjectDetailPage() {
  const { projectId = "" } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const {
    users,
    project,
    allTasks,
    tasks,
    loadingProject,
    loadingTasks,
    projectError,
    taskError,
    statusFilter,
    assigneeFilter,
    modalState,
    deletingTaskId,
    statusSavingId,
    visibleColumns,
    assigneeOptions,
    userMap,
    setStatusFilter,
    setAssigneeFilter,
    openCreateModal,
    openEditModal,
    closeModal
  } = useProjectDetailViewModel();
  const { refreshProjectAndTasks } = useProjectDetailQueryController(projectId);
  const { handleDeleteTask, handleStatusChange, handleTaskSaved } =
    useProjectDetailTaskController(projectId, refreshProjectAndTasks);

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
    <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-5">
      <ProjectDetailHeader
        allTaskCount={allTasks.length}
        description={project.description}
        name={project.name}
        roleLabel={canEditProject ? "Owner" : "Member"}
        visibleTaskCount={tasks.length}
      />

      <ProjectDetailFiltersCard
        assigneeFilter={assigneeFilter}
        assigneeOptions={assigneeOptions}
        currentUserId={user?.id}
        onAssigneeChange={setAssigneeFilter}
        onCreateTask={openCreateModal}
        onRefresh={() => void refreshProjectAndTasks()}
        onStatusChange={setStatusFilter}
        statusFilter={statusFilter}
        taskError={taskError}
        userMap={userMap}
      />

      <TaskBoard
        columns={visibleColumns}
        currentUserId={user?.id}
        deletingTaskId={deletingTaskId}
        loading={loadingTasks}
        onDeleteTask={(taskId) => void handleDeleteTask(taskId)}
        onEditTask={openEditModal}
        onStatusChange={(task, status) => void handleStatusChange(task, status as TaskStatus)}
        statusSavingId={statusSavingId}
        tasks={tasks}
        userMap={userMap}
      />

      {modalState ? (
        <TaskModal
          key={modalState.mode === "edit" ? modalState.task.id : "new-task"}
          mode={modalState.mode}
          onClose={closeModal}
          onSaved={() => void handleTaskSaved()}
          projectId={projectId}
          task={modalState.task}
          users={users}
        />
      ) : null}
    </main>
  );
}
