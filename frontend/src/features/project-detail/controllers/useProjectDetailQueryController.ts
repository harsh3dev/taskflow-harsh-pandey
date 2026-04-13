import { useEffect } from "react";
import { useShallow } from "zustand/react/shallow";
import { useApi } from "../../../app/auth";
import { getProject } from "../../../lib/services/projects";
import { listProjectTasks } from "../../../lib/services/tasks";
import { listUsers } from "../../../lib/services/users";
import { getErrorMessage } from "../../../lib/utils";
import { useProjectDetailStore } from "../store";

export function useProjectDetailQueryController(projectId: string) {
  const api = useApi();
  const {
    project,
    statusFilter,
    assigneeFilter,
    resetForProject,
    setShellData,
    setTasks,
    setLoadingProject,
    setLoadingTasks,
    setProjectError,
    setTaskError
  } = useProjectDetailStore(
    useShallow((state) => ({
      project: state.project,
      statusFilter: state.statusFilter,
      assigneeFilter: state.assigneeFilter,
      resetForProject: state.resetForProject,
      setShellData: state.setShellData,
      setTasks: state.setTasks,
      setLoadingProject: state.setLoadingProject,
      setLoadingTasks: state.setLoadingTasks,
      setProjectError: state.setProjectError,
      setTaskError: state.setTaskError
    }))
  );

  async function loadProjectShell() {
    setLoadingProject(true);
    setProjectError("");
    try {
      const [projectResponse, usersResponse] = await Promise.all([
        getProject(api, projectId),
        listUsers(api)
      ]);
      setShellData({
        project: projectResponse.project,
        tasks: projectResponse.tasks,
        users: usersResponse.users
      });
    } catch (error) {
      setProjectError(getErrorMessage(error, "Unable to load this project."));
    } finally {
      setLoadingProject(false);
    }
  }

  async function loadTasks(filters?: { status?: string; assignee?: string }) {
    setLoadingTasks(true);
    setTaskError("");
    try {
      const response = await listProjectTasks(api, projectId, filters);
      setTasks(response.tasks);
    } catch (error) {
      setTaskError(getErrorMessage(error, "Unable to refresh tasks."));
    } finally {
      setLoadingTasks(false);
    }
  }

  async function refreshProjectAndTasks() {
    const [projectResponse, usersResponse] = await Promise.all([
      getProject(api, projectId),
      listUsers(api)
    ]);
    setShellData({
      project: projectResponse.project,
      tasks: projectResponse.tasks,
      users: usersResponse.users
    });
    if (statusFilter || assigneeFilter) {
      await loadTasks({ status: statusFilter, assignee: assigneeFilter });
    }
  }

  useEffect(() => {
    resetForProject();
    void loadProjectShell();
  }, [projectId]);

  useEffect(() => {
    if (!project) {
      return;
    }
    void loadTasks({
      status: statusFilter,
      assignee: assigneeFilter
    });
  }, [project, statusFilter, assigneeFilter]);

  return {
    refreshProjectAndTasks
  };
}
