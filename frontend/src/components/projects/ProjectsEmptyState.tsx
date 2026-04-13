import { Card, CardContent } from "../ui/card";

export function ProjectsEmptyState() {
  return (
    <Card>
      <CardContent className="px-8 py-8 text-center">
        <h3>No projects yet</h3>
        <p className="text-[var(--ink-soft)]">Create your first project to start organizing tasks.</p>
      </CardContent>
    </Card>
  );
}
