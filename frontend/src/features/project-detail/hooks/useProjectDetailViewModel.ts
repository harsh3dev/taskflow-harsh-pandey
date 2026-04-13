import { useMemo } from "react";
import { useShallow } from "zustand/react/shallow";
import { statusOptions } from "../../../lib/constants";
import { useProjectDetailStore } from "../store";

export function useProjectDetailViewModel() {
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
    setStatusFilter,
    setAssigneeFilter,
    openCreateModal,
    openEditModal,
    closeModal
  } = useProjectDetailStore(
    useShallow((s) => ({
      users: s.users,
      project: s.project,
      allTasks: s.allTasks,
      tasks: s.tasks,
      loadingProject: s.loadingProject,
      loadingTasks: s.loadingTasks,
      projectError: s.projectError,
      taskError: s.taskError,
      statusFilter: s.statusFilter,
      assigneeFilter: s.assigneeFilter,
      modalState: s.modalState,
      deletingTaskId: s.deletingTaskId,
      statusSavingId: s.statusSavingId,
      setStatusFilter: s.setStatusFilter,
      setAssigneeFilter: s.setAssigneeFilter,
      openCreateModal: s.openCreateModal,
      openEditModal: s.openEditModal,
      closeModal: s.closeModal
    }))
  );

  const assigneeOptions = useMemo(
    () =>
      Array.from(
        new Set(
          allTasks
            .map((task) => task.assignee_id)
            .filter((value): value is string => Boolean(value))
        )
      ),
    [allTasks]
  );

  const visibleColumns = useMemo(
    () =>
      statusOptions.map((column) => ({
        ...column,
        tasks: tasks.filter((task) => task.status === column.value)
      })),
    [tasks]
  );

  const userMap = useMemo(
    () => new Map(users.map((entry) => [entry.id, entry])),
    [users]
  );

  return {
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
  };
}
