import { FormEvent } from "react";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { FormField, FormMessage } from "../ui/form-field";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { Textarea } from "../ui/textarea";

type CreateProjectCardProps = {
  name: string;
  description: string;
  errors: Record<string, string>;
  creating: boolean;
  onChange: (field: "name" | "description", value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function CreateProjectCard({
  name,
  description,
  errors,
  creating,
  onChange,
  onSubmit
}: CreateProjectCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Create a project</CardTitle>
        <CardDescription>Keep the scope tight and the description useful.</CardDescription>
      </CardHeader>

      <CardContent>
        <form className="flex flex-col gap-5" onSubmit={onSubmit}>
          <FormField>
            <Label htmlFor="project-name">Project name</Label>
            <Input
              id="project-name"
              value={name}
              onChange={(event) => onChange("name", event.target.value)}
              placeholder="Q2 Launch Readiness"
            />
            {errors.name ? <FormMessage>{errors.name}</FormMessage> : null}
          </FormField>

          <FormField>
            <Label htmlFor="project-description">Description</Label>
            <Textarea
              id="project-description"
              value={description}
              onChange={(event) => onChange("description", event.target.value)}
              placeholder="Goals, handoff notes, and delivery boundaries."
            />
          </FormField>

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <Button disabled={creating} type="submit">
              {creating ? "Creating..." : "Create project"}
            </Button>
            <span className="text-sm text-muted-foreground">
              Projects are visible if you own them or have tasks in them.
            </span>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
