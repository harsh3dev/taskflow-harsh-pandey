import { HTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/utils";

export function FormField({
  className,
  children,
  ...props
}: HTMLAttributes<HTMLDivElement> & { children: ReactNode }) {
  return (
    <div className={cn("ui-form-field grid gap-1.5", className)} {...props}>
      {children}
    </div>
  );
}

export function FormMessage({
  className,
  ...props
}: HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("text-sm text-destructive", className)} {...props} />;
}

export function FormDescription({
  className,
  ...props
}: HTMLAttributes<HTMLParagraphElement>) {
  return <p className={cn("text-sm text-muted-foreground", className)} {...props} />;
}
