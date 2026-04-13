import { FormEvent, useState } from "react";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { FormField, FormMessage } from "../ui/form-field";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";

type CreateProjectModalProps = {
  creating: boolean;
  errors: Record<string, string>;
  onSubmit: (name: string, description: string) => Promise<void>;
  onClose: () => void;
};

export function CreateProjectModal({
  creating,
  errors,
  onSubmit,
  onClose
}: CreateProjectModalProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    await onSubmit(name, description);
  }

  return (
    <div
      className="fixed inset-0 z-20 grid place-items-center bg-black/50 p-4"
      role="presentation"
    >
      <Card className="w-full max-w-[520px] rounded-xl shadow-2xl sm:rounded-2xl">
        <CardHeader className="flex flex-row items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-primary">
              New project
            </p>
            <CardTitle>Create a project</CardTitle>
            <CardDescription>Keep the scope tight and the description useful.</CardDescription>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose} type="button">
            Close
          </Button>
        </CardHeader>
        <CardContent>
          <form className="flex flex-col gap-5" onSubmit={handleSubmit}>
            <FormField>
              <Label htmlFor="new-project-name">Project name</Label>
              <Input
                id="new-project-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Q2 Launch Readiness"
                autoFocus
              />
              {errors.name ? <FormMessage>{errors.name}</FormMessage> : null}
            </FormField>

            <FormField>
              <Label htmlFor="new-project-desc">Description</Label>
              <Textarea
                id="new-project-desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Goals, handoff notes, and delivery boundaries."
              />
            </FormField>

            <div className="flex gap-3">
              <Button disabled={creating} type="submit">
                {creating ? "Creating…" : "Create project"}
              </Button>
              <Button variant="outline" onClick={onClose} type="button">
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
