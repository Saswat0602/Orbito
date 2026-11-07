import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 transition-colors focus:outline-none focus:ring-2 focus:ring-zinc-400 focus:ring-offset-2 [&>svg]:size-3 gap-1",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-zinc-100 text-zinc-950 shadow-xs hover:bg-zinc-200",
        secondary:
          "border-transparent bg-zinc-800 text-zinc-200 hover:bg-zinc-750",
        destructive:
          "border-transparent bg-rose-950/60 border border-rose-900/50 text-rose-300 hover:bg-rose-900/50",
        outline:
          "text-zinc-300 border-zinc-700 hover:bg-zinc-800 hover:text-zinc-100",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

function Badge({
  className,
  variant,
  asChild = false,
  ...props
}: React.ComponentProps<"span"> &
  VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
  const Comp = asChild ? Slot : "span";

  return (
    <Comp
      data-slot="badge"
      className={cn(badgeVariants({ variant }), className)}
      {...props}
    />
  );
}

export { Badge, badgeVariants };
