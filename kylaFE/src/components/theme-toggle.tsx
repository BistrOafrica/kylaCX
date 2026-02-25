import { useTheme } from "./theme-provider";
import { MoonIcon, SunIcon } from "lucide-react";
import { Button } from "./ui/button";
import { cn } from "@/lib/utils";

export default function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <Button
      data-sidebar="trigger"
      data-slot="sidebar-trigger"
      variant="ghost"
      size="icon"
      className={cn("size-7")}
      onClick={() => {
        setTheme(theme === "light" ? "dark" : "light");
      }}
    >
      {theme === "light" ? (
        <MoonIcon
          className="text-foreground"
          onClick={() => setTheme("dark")}
        />
      ) : (
        <SunIcon
          className="text-foreground"
          onClick={() => setTheme("light")}
        />
      )}
    </Button>
    // <Toggle aria-label="Toggle bookmark" className="">

    // </Toggle>
  );
}
