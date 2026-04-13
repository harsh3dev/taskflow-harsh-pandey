import { Component, ErrorInfo, ReactNode } from "react";
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";

export class AppErrorBoundary extends Component<
  { children: ReactNode },
  { hasError: boolean }
> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  componentDidCatch(_error: Error, _errorInfo: ErrorInfo) {}

  render() {
    if (this.state.hasError) {
      return (
        <main className="mx-auto flex w-full max-w-6xl px-4 py-6 sm:px-5">
          <Card className="mx-auto mt-12 w-full max-w-3xl">
            <CardContent className="flex flex-col items-center gap-4 px-8 py-10 text-center">
              <p className="text-xs font-semibold uppercase tracking-[0.24em] text-primary">
                Application error
              </p>
            <h1>TaskFlow hit an unexpected problem.</h1>
            <p>Refresh the page and try again. Your saved session will be kept in local storage.</p>
            <Button onClick={() => window.location.reload()} type="button">
              Reload app
            </Button>
            </CardContent>
          </Card>
        </main>
      );
    }

    return this.props.children;
  }
}
