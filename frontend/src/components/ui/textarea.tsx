import { TextareaHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export const Textarea = forwardRef<
  HTMLTextAreaElement,
  TextareaHTMLAttributes<HTMLTextAreaElement>
>(function Textarea({ className, ...props }, ref) {
  return (
    <textarea
      ref={ref}
      className={cn(
        "flex min-h-[120px] w-full rounded-2xl border border-[var(--line-strong)] bg-white/80 px-4 py-3 text-sm text-[var(--ink)] transition-[border-color,box-shadow,transform] outline-none placeholder:text-[var(--ink-soft)]/70 focus-visible:border-[rgba(201,109,66,0.75)] focus-visible:ring-4 focus-visible:ring-[rgba(201,109,66,0.14)] disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    />
  );
});
