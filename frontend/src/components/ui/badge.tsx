import { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

type BadgeVariant = "default" | "secondary" | "outline";

export function Badge({
  className,
  variant = "default",
  ...props
}: HTMLAttributes<HTMLSpanElement> & { variant?: BadgeVariant }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full bg-panel/10 px-[0.65rem] py-[0.36rem] text-[0.78rem] font-medium text-panel",
        variant === "secondary" && "bg-panel/[0.08]",
        variant === "outline" && "border border-border bg-transparent text-foreground",
        className
      )}
      {...props}
    />
  );
}
