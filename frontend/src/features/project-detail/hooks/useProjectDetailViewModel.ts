import { statusOptions } from "../../../lib/constants";
import { useProjectDetailStore } from "../store";

export function useProjectDetailViewModel() {
  return useProjectDetailStore((state) => {
    const assigneeOptions = Array.from(
      new Set(
        state.allTasks
          .map((task) => task.assignee_id)
          .filter((value): value is string => Boolean(value))
      )
    );

    const visibleColumns = statusOptions.map((column) => ({
      ...column,
      tasks: state.tasks.filter((task) => task.status === column.value)
    }));

    return {
      users: state.users,
      project: state.project,
      allTasks: state.allTasks,
      tasks: state.tasks,
      loadingProject: state.loadingProject,
      loadingTasks: state.loadingTasks,
      projectError: state.projectError,
      taskError: state.taskError,
      statusFilter: state.statusFilter,
      assigneeFilter: state.assigneeFilter,
      modalState: state.modalState,
      deletingTaskId: state.deletingTaskId,
      statusSavingId: state.statusSavingId,
      visibleColumns,
      assigneeOptions,
      userMap: new Map(state.users.map((entry) => [entry.id, entry])),
      setStatusFilter: state.setStatusFilter,
      setAssigneeFilter: state.setAssigneeFilter,
      openCreateModal: state.openCreateModal,
      openEditModal: state.openEditModal,
      closeModal: state.closeModal
    };
  });
}
