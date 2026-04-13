import { ReactNode } from "react";
import { Alert, AlertDescription } from "../ui/alert";
import { Button } from "../ui/button";
import { Card, CardContent } from "../ui/card";

type ProjectDetailStateProps = {
  message: string;
  actionLabel?: string;
  onAction?: () => void;
  tone?: "default" | "destructive";
  children?: ReactNode;
};

export function ProjectDetailState({
  message,
  actionLabel,
  onAction,
  tone = "default",
  children
}: ProjectDetailStateProps) {
  return (
    <main className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-6 sm:px-5">
      <Card className="mx-auto mt-16 w-full max-w-[34rem]">
        <CardContent className="flex flex-col gap-4">
          <Alert variant={tone === "destructive" ? "destructive" : "default"}>
            <AlertDescription>{message}</AlertDescription>
          </Alert>
          {children}
          {actionLabel && onAction ? (
            <div>
              <Button
                variant={tone === "destructive" ? "secondary" : "outline"}
                onClick={onAction}
                type="button"
              >
                {actionLabel}
              </Button>
            </div>
          ) : null}
        </CardContent>
      </Card>
    </main>
  );
}
