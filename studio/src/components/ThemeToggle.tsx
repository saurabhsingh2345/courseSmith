// Light/dark switch.
//
// The machinery for this already existed — `applyTheme` resolves every token
// per mode, `preferredMode` reads the stored choice and falls back to the OS —
// and nothing in the UI had ever called it. The studio booted in whatever mode
// the OS asked for and could not be moved off it.

import { Moon, Sun } from "lucide-react";
import { Button } from "./base/Button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "./base/Tooltip";
import { studio, useStudioStore } from "../store/studioStore";

export function ThemeToggle() {
  const theme = useStudioStore((s) => s.theme);
  const next = theme === "dark" ? "light" : "dark";
  const label = `Switch to ${next} mode`;

  return (
    <TooltipProvider delayDuration={200}>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            onClick={studio.toggleTheme}
            aria-label={label}
            className="size-11 sm:size-10"
          >
            {theme === "dark" ? (
              <Sun className="size-4" aria-hidden />
            ) : (
              <Moon className="size-4" aria-hidden />
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">{label}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
