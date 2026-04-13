import { Card, CardContent } from "../ui/card";

type ProjectsPageHeaderProps = {
  projectCount: number;
  ownedCount: number;
};

export function ProjectsPageHeader({
  projectCount,
  ownedCount
}: ProjectsPageHeaderProps) {
  return (
    <section className="flex flex-col gap-4 rounded-[30px] border border-white/40 bg-[rgba(255,251,246,0.6)] p-6 backdrop-blur-sm sm:p-8">
      <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[var(--accent-strong)]">
        Workspace
      </p>
      <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
        <div className="flex max-w-2xl flex-col gap-3">
          <h1>Projects that matter today.</h1>
          <p className="text-[var(--ink-soft)]">
            Create a project, keep a clean overview of your accessible workspaces, and jump
            straight into grouped task management.
          </p>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:w-[320px]">
          <Card className="rounded-[24px] bg-[rgba(255,251,246,0.78)] backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-[rgba(19,59,51,0.08)] px-[0.65rem] py-[0.36rem] text-[0.78rem] text-[var(--panel)]">
                Total projects
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {projectCount}
              </strong>
            </CardContent>
          </Card>
          <Card className="rounded-[24px] bg-[rgba(255,251,246,0.78)] backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-[rgba(19,59,51,0.08)] px-[0.65rem] py-[0.36rem] text-[0.78rem] text-[var(--panel)]">
                Owned by you
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {ownedCount}
              </strong>
            </CardContent>
          </Card>
        </div>
      </div>
    </section>
  );
}
