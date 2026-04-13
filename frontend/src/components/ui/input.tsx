import { InputHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils";

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  function Input({ className, ...props }, ref) {
    return (
      <input
        ref={ref}
        className={cn(
          "flex h-12 w-full rounded-2xl border border-[var(--line-strong)] bg-white/80 px-4 py-3 text-sm text-[var(--ink)] transition-[border-color,box-shadow,transform] outline-none placeholder:text-[var(--ink-soft)]/70 focus-visible:border-[rgba(201,109,66,0.75)] focus-visible:ring-4 focus-visible:ring-[rgba(201,109,66,0.14)] disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        {...props}
      />
    );
  }
);
