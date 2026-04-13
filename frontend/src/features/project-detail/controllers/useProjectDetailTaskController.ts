import { useApi } from "../../../app/auth";
import { deleteTask, updateTask } from "../../../lib/services/tasks";
import { getErrorMessage } from "../../../lib/utils";
import { Task, TaskStatus } from "../../../lib/types";
import { useShallow } from "zustand/react/shallow";
import { useProjectDetailStore } from "../store";

export function useProjectDetailTaskController(
  projectId: string,
  onRefresh: () => Promise<void>
) {
  const api = useApi();
  const {
    tasks,
    allTasks,
    setDeletingTaskId,
    setTaskError,
    removeTask,
    setStatusSavingId,
    applyOptimisticStatus,
    replaceTask,
    closeModal
  } = useProjectDetailStore(
    useShallow((state) => ({
      tasks: state.tasks,
      allTasks: state.allTasks,
      setDeletingTaskId: state.setDeletingTaskId,
      setTaskError: state.setTaskError,
      removeTask: state.removeTask,
      setStatusSavingId: state.setStatusSavingId,
      applyOptimisticStatus: state.applyOptimisticStatus,
      replaceTask: state.replaceTask,
      closeModal: state.closeModal
    }))
  );

  async function handleDeleteTask(taskId: string) {
    setDeletingTaskId(taskId);
    setTaskError("");
    try {
      await deleteTask(api, taskId);
      removeTask(taskId);
    } catch (error) {
      setTaskError(getErrorMessage(error, "Unable to delete the task."));
    } finally {
      setDeletingTaskId(null);
    }
  }

  async function handleStatusChange(task: Task, status: TaskStatus) {
    const previousTasks = tasks;
    const previousAllTasks = allTasks;

    setStatusSavingId(task.id);
    setTaskError("");
    applyOptimisticStatus(task.id, status);

    try {
      const response = await updateTask(api, task.id, { status });
      replaceTask(response.task);
    } catch (error) {
      useProjectDetailStore.setState({
        tasks: previousTasks,
        allTasks: previousAllTasks
      });
      setTaskError(getErrorMessage(error, "Unable to update task status."));
    } finally {
      setStatusSavingId(null);
    }
  }

  async function handleTaskSaved() {
    closeModal();
    try {
      await onRefresh();
    } catch (error) {
      setTaskError(getErrorMessage(error, "Task saved, but the refreshed view failed."));
    }
  }

  return {
    handleDeleteTask,
    handleStatusChange,
    handleTaskSaved
  };
}
