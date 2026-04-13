import { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type AlertVariant = "default" | "destructive";

export function Alert({
  className,
  variant = "default",
  ...props
}: HTMLAttributes<HTMLDivElement> & { variant?: AlertVariant }) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-[var(--line)] bg-[rgba(19,59,51,0.04)] px-4 py-3",
        variant === "destructive" &&
          "border-[rgba(177,69,62,0.24)] bg-[rgba(177,69,62,0.08)] text-[var(--danger)]",
        className
      )}
      {...props}
    />
  );
}

export function AlertTitle({ className, ...props }: HTMLAttributes<HTMLHeadingElement>) {
  return <h4 className={cn("mb-1 font-semibold leading-none tracking-tight", className)} {...props} />;
}

export function AlertDescription({ className, ...props }: HTMLAttributes<HTMLParagraphElement>) {
  return <div className={cn("text-sm text-[var(--ink-soft)]", className)} {...props} />;
}
