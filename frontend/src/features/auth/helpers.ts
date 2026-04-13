export type AuthMode = "login" | "register";

export function validateAuthForm(
  mode: AuthMode,
  form: { name: string; email: string; password: string }
) {
  const nextErrors: Record<string, string> = {};
  if (mode === "register" && form.name.trim().length < 2) {
    nextErrors.name = "Enter at least 2 characters";
  }
  if (!form.email.trim()) {
    nextErrors.email = "Email is required";
  } else if (!/\S+@\S+\.\S+/.test(form.email.trim())) {
    nextErrors.email = "Enter a valid email";
  }
  if (form.password.trim().length < 8) {
    nextErrors.password = "Use at least 8 characters";
  }
  return nextErrors;
}

export function getAuthCopy(mode: AuthMode) {
  if (mode === "login") {
    return {
      title: "Welcome back.",
      subtitle: "Sign in to manage projects, track tasks, and pick up where you left off.",
      heading: "Sign in to TaskFlow",
      eyebrow: "Login",
      helper: "Use the seeded credentials or your own account.",
      cta: "Sign in",
      pendingCta: "Signing in...",
      swapLead: "Need an account? ",
      swapHref: "/register",
      swapLabel: "Create one"
    };
  }

  return {
    title: "Build your workspace.",
    subtitle: "Create an account, save your session, and move straight into project planning.",
    heading: "Create your account",
    eyebrow: "Register",
    helper: "Your account will be logged in immediately after registration.",
    cta: "Create account",
    pendingCta: "Creating account...",
    swapLead: "Already registered? ",
    swapHref: "/login",
    swapLabel: "Sign in"
  };
}
