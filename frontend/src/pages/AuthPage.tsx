import { FormEvent, useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../app/auth";
import { ApiError } from "../lib/api";
import { AuthFormCard } from "../components/auth/AuthFormCard";
import { AuthHero } from "../components/auth/AuthHero";
import { getAuthCopy, validateAuthForm, type AuthMode } from "../features/auth/helpers";

type AuthPageProps = {
  mode: AuthMode;
};

export function AuthPage({ mode }: AuthPageProps) {
  const auth = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const nextPath = (location.state as { from?: string } | null)?.from || "/";
  const [form, setForm] = useState({
    name: "",
    email: "",
    password: ""
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [errorMessage, setErrorMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (auth.token && auth.user) {
      navigate(nextPath, { replace: true });
    }
  }, [auth.token, auth.user, navigate, nextPath]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const nextErrors = validateAuthForm(mode, form);
    setFieldErrors(nextErrors);
    setErrorMessage("");

    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    setSubmitting(true);
    try {
      if (mode === "login") {
        await auth.login({
          email: form.email.trim(),
          password: form.password
        });
      } else {
        await auth.register({
          name: form.name.trim(),
          email: form.email.trim(),
          password: form.password
        });
      }
      navigate(nextPath, { replace: true });
    } catch (error) {
      if (error instanceof ApiError) {
        setFieldErrors(error.fields ?? {});
        setErrorMessage(error.message);
      } else {
        setErrorMessage("Unable to reach the API.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  const copy = getAuthCopy(mode);

  return (
    <main className="mx-auto grid min-h-screen w-full max-w-[1200px] grid-cols-1 gap-6 px-4 py-6 lg:grid-cols-[1.15fr_0.85fr] lg:items-center lg:px-6">
      <AuthHero title={copy.title} subtitle={copy.subtitle} />
      <AuthFormCard
        copy={copy}
        errorMessage={errorMessage}
        fieldErrors={fieldErrors}
        form={form}
        mode={mode}
        submitting={submitting}
        onChange={(field, value) =>
          setForm((current) => ({
            ...current,
            [field]: value
          }))
        }
        onSubmit={handleSubmit}
      />
    </main>
  );
}
