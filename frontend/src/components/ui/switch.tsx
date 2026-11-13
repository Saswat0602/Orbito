"use client";

import * as React from "react";
import * as SwitchPrimitive from "@radix-ui/react-switch";
import { cn } from "@/lib/utils";

function Switch({ className, ...props }: React.ComponentProps<typeof SwitchPrimitive.Root>) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      className={cn(
        "peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border transition-all duration-200 outline-none",
        "data-[state=checked]:bg-zinc-100 data-[state=checked]:border-white shadow-xs",
        "data-[state=unchecked]:bg-zinc-900 data-[state=unchecked]:border-zinc-750 hover:data-[state=unchecked]:border-zinc-600",
        "focus-visible:ring-2 focus-visible:ring-zinc-400 focus-visible:ring-offset-2 focus-visible:ring-offset-[#09090b]",
        "disabled:cursor-not-allowed disabled:opacity-40",
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        data-slot="switch-thumb"
        className={cn(
          "pointer-events-none block h-3.5 w-3.5 rounded-full transition-transform duration-200 shadow-sm",
          "data-[state=checked]:bg-zinc-950 data-[state=checked]:translate-x-[18px]",
          "data-[state=unchecked]:bg-zinc-400 data-[state=unchecked]:translate-x-[2px]"
        )}
      />
    </SwitchPrimitive.Root>
  );
}

export { Switch };
