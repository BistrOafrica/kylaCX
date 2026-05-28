import { useState } from "react";
import { Button } from "@/components/ui/button";
import { IconPhone } from "@tabler/icons-react";
import { CallWidget } from "./call-widget";

export function FloatingCallButton() {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      {/* Floating Action Button */}
      <Button
        size="icon"
        className="fixed bottom-6 right-6 size-16 rounded-full shadow-2xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-700 hover:to-indigo-700 z-50 transition-all hover:scale-110"
        onClick={() => setIsOpen(!isOpen)}
      >
        <IconPhone className="size-6" />
      </Button>

      {/* Call Widget */}
      {isOpen && (
        <div className="fixed bottom-28 right-6 z-50">
          <CallWidget onClose={() => setIsOpen(false)} />
        </div>
      )}
    </>
  );
}
