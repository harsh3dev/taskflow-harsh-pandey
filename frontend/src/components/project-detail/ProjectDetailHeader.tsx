import { Link } from "react-router-dom";
import { Card, CardContent } from "../ui/card";
import { ProjectStats } from "../../lib/types";

type ProjectDetailHeaderProps = {
  name: string;
  description: string;
  visibleTaskCount: number;
  allTaskCount: number;
  roleLabel: string;
  stats?: ProjectStats | null;
};

export function ProjectDetailHeader({
  name,
  description,
  visibleTaskCount,
  allTaskCount,
  roleLabel,
  stats
}: ProjectDetailHeaderProps) {
  return (
    <section className="flex flex-col gap-4 rounded-[30px] border border-border/40 bg-card/60 p-6 backdrop-blur-sm sm:p-8">
      <div className="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
        <div className="flex max-w-2xl flex-col gap-3">
          <Link className="font-bold text-primary" to="/">
            ← Back to projects
          </Link>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-primary">
            Project detail
          </p>
          <h1>{name}</h1>
          <p className="text-muted-foreground">{description || "No project description yet."}</p>
        </div>
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 xl:w-[520px]">
          <Card className="rounded-[24px] bg-card/80 backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-panel/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] text-panel">
                Visible
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {visibleTaskCount}
              </strong>
            </CardContent>
          </Card>
          <Card className="rounded-[24px] bg-card/80 backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-panel/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] text-panel">
                Total
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {allTaskCount}
              </strong>
            </CardContent>
          </Card>
          <Card className="rounded-[24px] bg-card/80 backdrop-blur-sm">
            <CardContent className="p-5">
              <span className="inline-flex rounded-full bg-panel/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] text-panel">
                Role
              </span>
              <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                {roleLabel}
              </strong>
            </CardContent>
          </Card>
          {stats ? (
            <>
              <Card className="rounded-[24px] bg-card/80 backdrop-blur-sm">
                <CardContent className="p-5">
                  <span className="inline-flex rounded-full bg-success/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] text-success">
                    Done
                  </span>
                  <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                    {stats.status_counts.done ?? 0}
                  </strong>
                </CardContent>
              </Card>
              <Card className="rounded-[24px] bg-card/80 backdrop-blur-sm">
                <CardContent className="p-5">
                  <span className="inline-flex rounded-full bg-warning/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] text-warning">
                    In progress
                  </span>
                  <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                    {stats.status_counts.in_progress ?? 0}
                  </strong>
                </CardContent>
              </Card>
              <Card className="rounded-[24px] bg-card/80 backdrop-blur-sm">
                <CardContent className="p-5">
                  <span className="inline-flex rounded-full bg-panel/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] text-panel">
                    To do
                  </span>
                  <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                    {stats.status_counts.todo ?? 0}
                  </strong>
                </CardContent>
              </Card>
            </>
          ) : null}
        </div>
      </div>
    </section>
  );
}
