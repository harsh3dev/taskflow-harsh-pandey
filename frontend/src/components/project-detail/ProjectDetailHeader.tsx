import { Link } from "react-router-dom";
import { Card, CardContent } from "../ui/card";

type ProjectDetailHeaderProps = {
  name: string;
  description: string;
  visibleTaskCount: number;
  allTaskCount: number;
  roleLabel: string;
};

export function ProjectDetailHeader({
  name,
  description,
  visibleTaskCount,
  allTaskCount,
  roleLabel
}: ProjectDetailHeaderProps) {
  return (
    <section className="flex flex-col gap-4 rounded-[30px] border border-white/40 bg-[rgba(255,251,246,0.6)] p-6 backdrop-blur-sm sm:p-8">
      <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
        <div className="flex max-w-2xl flex-col gap-3">
          <Link className="font-bold text-[var(--accent-strong)]" to="/">
            ← Back to projects
          </Link>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[var(--accent-strong)]">
            Project detail
          </p>
          <h1>{name}</h1>
          <p className="text-[var(--ink-soft)]">{description || "No project description yet."}</p>
        </div>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3 xl:w-[520px]">
          <Card className="rounded-[24px] bg-[rgba(255,251,246,0.78)] backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-[rgba(19,59,51,0.08)] px-[0.65rem] py-[0.36rem] text-[0.78rem] text-[var(--panel)]">
                Visible tasks
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {visibleTaskCount}
              </strong>
            </CardContent>
          </Card>
          <Card className="rounded-[24px] bg-[rgba(255,251,246,0.78)] backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-[rgba(19,59,51,0.08)] px-[0.65rem] py-[0.36rem] text-[0.78rem] text-[var(--panel)]">
                All tasks
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {allTaskCount}
              </strong>
            </CardContent>
          </Card>
          <Card className="rounded-[24px] bg-[rgba(255,251,246,0.78)] backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-[rgba(19,59,51,0.08)] px-[0.65rem] py-[0.36rem] text-[0.78rem] text-[var(--panel)]">
                Role
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {roleLabel}
              </strong>
            </CardContent>
          </Card>
        </div>
      </div>
    </section>
  );
}
