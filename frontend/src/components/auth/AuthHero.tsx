import { Card, CardContent } from "../ui/card";

type AuthHeroProps = {
  title: string;
  subtitle: string;
};

const stats = [
  { label: "Protected routes", value: "JWT" },
  { label: "Project views", value: "Kanban" },
  { label: "Task updates", value: "Optimistic" },
  { label: "Session", value: "Persistent" }
];

export function AuthHero({ title, subtitle }: AuthHeroProps) {
  return (
    <section className="px-1 py-4 lg:px-2 lg:py-6">
      <div className="flex flex-col gap-6">
        <div className="rounded-[28px] border border-white/40 bg-[linear-gradient(145deg,rgba(19,59,51,0.98),rgba(36,86,76,0.9))] p-8 text-[#f9f4ec] shadow-[0_24px_80px_rgba(22,33,30,0.18)] sm:p-10">
          <div className="flex flex-col gap-5">
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[#d6bfa5]">
              Phase 3 Frontend
            </p>
            <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">{title}</h1>
            <p className="max-w-2xl text-base leading-7 text-[#f1e6d7]/82">{subtitle}</p>
          </div>
        </div>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {stats.map((stat) => (
            <Card
              className="rounded-[24px] bg-[rgba(255,251,246,0.78)] backdrop-blur-sm"
              key={stat.label}
            >
              <CardContent className="p-5">
                <span className="inline-flex rounded-full bg-[rgba(19,59,51,0.08)] px-[0.65rem] py-[0.36rem] text-[0.78rem] text-[var(--panel)]">
                  {stat.label}
                </span>
                <strong className="mt-3 block text-[2rem] font-semibold tracking-tight">
                  {stat.value}
                </strong>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}
