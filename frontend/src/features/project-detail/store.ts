import { create } from "zustand";
import { Project, Task, TaskStatus, User } from "../../lib/types";

export type TaskModalState =
  | { mode: "create"; task: null }
  | { mode: "edit"; task: Task }
  | undefined;

type ProjectDetailState = {
  users: User[];
  project: Project | null;
  allTasks: Task[];
  tasks: Task[];
  loadingProject: boolean;
  loadingTasks: boolean;
  projectError: string;
  taskError: string;
  statusFilter: string;
  assigneeFilter: string;
  modalState: TaskModalState;
  deletingTaskId: string | null;
  statusSavingId: string | null;
  resetForProject: () => void;
  setShellData: (payload: { project: Project; tasks: Task[]; users: User[] }) => void;
  setTasks: (tasks: Task[]) => void;
  setLoadingProject: (value: boolean) => void;
  setLoadingTasks: (value: boolean) => void;
  setProjectError: (value: string) => void;
  setTaskError: (value: string) => void;
  setStatusFilter: (value: string) => void;
  setAssigneeFilter: (value: string) => void;
  openCreateModal: () => void;
  openEditModal: (task: Task) => void;
  closeModal: () => void;
  setDeletingTaskId: (value: string | null) => void;
  setStatusSavingId: (value: string | null) => void;
  removeTask: (taskId: string) => void;
  replaceTask: (task: Task) => void;
  applyOptimisticStatus: (taskId: string, status: TaskStatus) => void;
};

const initialState = {
  users: [] as User[],
  project: null as Project | null,
  allTasks: [] as Task[],
  tasks: [] as Task[],
  loadingProject: true,
  loadingTasks: false,
  projectError: "",
  taskError: "",
  statusFilter: "",
  assigneeFilter: "",
  modalState: undefined as TaskModalState,
  deletingTaskId: null as string | null,
  statusSavingId: null as string | null
};

export const useProjectDetailStore = create<ProjectDetailState>((set) => ({
  ...initialState,
  resetForProject: () => set({ ...initialState }),
  setShellData: ({ project, tasks, users }) =>
    set({
      project,
      allTasks: tasks,
      tasks,
      users
    }),
  setTasks: (tasks) => set({ tasks }),
  setLoadingProject: (value) => set({ loadingProject: value }),
  setLoadingTasks: (value) => set({ loadingTasks: value }),
  setProjectError: (value) => set({ projectError: value }),
  setTaskError: (value) => set({ taskError: value }),
  setStatusFilter: (value) => set({ statusFilter: value }),
  setAssigneeFilter: (value) => set({ assigneeFilter: value }),
  openCreateModal: () => set({ modalState: { mode: "create", task: null } }),
  openEditModal: (task) => set({ modalState: { mode: "edit", task } }),
  closeModal: () => set({ modalState: undefined }),
  setDeletingTaskId: (value) => set({ deletingTaskId: value }),
  setStatusSavingId: (value) => set({ statusSavingId: value }),
  removeTask: (taskId) =>
    set((state) => ({
      tasks: state.tasks.filter((task) => task.id !== taskId),
      allTasks: state.allTasks.filter((task) => task.id !== taskId)
    })),
  replaceTask: (task) =>
    set((state) => ({
      tasks: state.tasks.map((item) => (item.id === task.id ? task : item)),
      allTasks: state.allTasks.map((item) => (item.id === task.id ? task : item))
    })),
  applyOptimisticStatus: (taskId, status) =>
    set((state) => ({
      tasks: state.tasks.map((item) => (item.id === taskId ? { ...item, status } : item)),
      allTasks: state.allTasks.map((item) =>
        item.id === taskId ? { ...item, status } : item
      )
    }))
}));
