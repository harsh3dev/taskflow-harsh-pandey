import { FormEvent } from "react";
import { Link } from "react-router-dom";
import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../ui/card";
import { FormDescription, FormField, FormMessage } from "../ui/form-field";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import { AuthMode } from "../../features/auth/helpers";

type AuthFormCardProps = {
  mode: AuthMode;
  form: { name: string; email: string; password: string };
  fieldErrors: Record<string, string>;
  errorMessage: string;
  submitting: boolean;
  copy: {
    eyebrow: string;
    heading: string;
    helper: string;
    cta: string;
    pendingCta: string;
    swapLead: string;
    swapHref: string;
    swapLabel: string;
  };
  onChange: (field: "name" | "email" | "password", value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function AuthFormCard({
  mode,
  form,
  fieldErrors,
  errorMessage,
  submitting,
  copy,
  onChange,
  onSubmit
}: AuthFormCardProps) {
  return (
    <Card className="mx-auto w-full max-w-[540px] p-2 sm:p-4">
      <CardHeader>
        <p className="text-xs font-semibold uppercase tracking-[0.24em] text-[var(--accent-strong)]">
          {copy.eyebrow}
        </p>
        <CardTitle>{copy.heading}</CardTitle>
        <CardDescription>{copy.helper}</CardDescription>
      </CardHeader>

      <CardContent className="flex flex-col gap-5">
        <form className="flex flex-col gap-5" onSubmit={onSubmit}>
          {errorMessage ? (
            <Alert variant="destructive">
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          ) : null}

          {mode === "register" ? (
            <FormField>
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                name="name"
                value={form.name}
                onChange={(event) => onChange("name", event.target.value)}
                placeholder="Avery Chen"
              />
              {fieldErrors.name ? <FormMessage>{fieldErrors.name}</FormMessage> : null}
            </FormField>
          ) : null}

          <FormField>
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              name="email"
              type="email"
              value={form.email}
              onChange={(event) => onChange("email", event.target.value)}
              placeholder="test@example.com"
              autoComplete="email"
            />
            {fieldErrors.email ? <FormMessage>{fieldErrors.email}</FormMessage> : null}
          </FormField>

          <FormField>
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              name="password"
              type="password"
              value={form.password}
              onChange={(event) => onChange("password", event.target.value)}
              placeholder="password123"
              autoComplete={mode === "login" ? "current-password" : "new-password"}
            />
            {fieldErrors.password ? (
              <FormMessage>{fieldErrors.password}</FormMessage>
            ) : (
              <FormDescription>Minimum length: 8 characters.</FormDescription>
            )}
          </FormField>

          <Button disabled={submitting} type="submit">
            {submitting ? copy.pendingCta : copy.cta}
          </Button>
        </form>

        <p className="text-sm text-[var(--ink-soft)]">
          {copy.swapLead}
          <Link className="ml-1 font-bold text-[var(--accent-strong)]" to={copy.swapHref}>
            {copy.swapLabel}
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
