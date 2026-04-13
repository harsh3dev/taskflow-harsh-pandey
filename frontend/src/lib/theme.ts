const THEME_KEY = "taskflow-theme";

export type Theme = "light" | "dark";

export function getStoredTheme(): Theme {
  return (localStorage.getItem(THEME_KEY) as Theme) ?? "light";
}

export function applyTheme(theme: Theme) {
  if (theme === "dark") {
    document.documentElement.classList.add("dark");
  } else {
    document.documentElement.classList.remove("dark");
  }
  localStorage.setItem(THEME_KEY, theme);
}

export function toggleTheme(current: Theme): Theme {
  return current === "dark" ? "light" : "dark";
}
