import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import {
  IconPhone,
  IconPhoneOff,
  IconVolume,
  IconMicrophone,
  IconMicrophoneOff,
  IconChevronDown,
  IconChevronUp,
  IconBackspace,
  IconPlus,
  IconX,
} from "@tabler/icons-react";
import { Card } from "@/components/ui/card";

type CallState = "idle" | "calling" | "active" | "incoming";

interface CallWidgetProps {
  onClose: () => void;
}

export function CallWidget({ onClose }: CallWidgetProps) {
  const [phoneNumber, setPhoneNumber] = useState("");
  const [callState, setCallState] = useState<CallState>("idle");
  const [isMuted, setIsMuted] = useState(false);
  const [isSpeakerOn, setIsSpeakerOn] = useState(false);
  const [showDialpad, setShowDialpad] = useState(true);
  const [callDuration, setCallDuration] = useState(0);
  const [mergedCalls, setMergedCalls] = useState<string[]>([]);

  const dialpadKeys = [
    ["1", "2", "3"],
    ["4", "5", "6"],
    ["7", "8", "9"],
    ["*", "0", "#"],
  ];

  const handleDialpadClick = (key: string) => {
    setPhoneNumber((prev) => prev + key);
  };

  const handleBackspace = () => {
    setPhoneNumber((prev) => prev.slice(0, -1));
  };

  const handleCall = () => {
    if (phoneNumber.length === 0) return;
    
    setCallState("calling");
    // Simulate call connection
    setTimeout(() => {
      setCallState("active");
      startCallTimer();
    }, 2000);
  };

  const handleEndCall = () => {
    setCallState("idle");
    setCallDuration(0);
    setIsMuted(false);
    setIsSpeakerOn(false);
    setMergedCalls([]);
  };

  const handleMergeCall = () => {
    if (phoneNumber) {
      setMergedCalls((prev) => [...prev, phoneNumber]);
      setPhoneNumber("");
    }
  };

  const startCallTimer = () => {
    const interval = setInterval(() => {
      setCallDuration((prev) => prev + 1);
    }, 1000);
    return () => clearInterval(interval);
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <Card className="w-96 shadow-2xl border-2 animate-in slide-in-from-bottom-5 duration-300">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30">
        <div className="flex-1">
          <h3 className="font-semibold text-lg">
            {callState === "idle" && "Dialer"}
            {callState === "calling" && "Calling..."}
            {callState === "active" && "In Call"}
            {callState === "incoming" && "Incoming Call"}
          </h3>
          {callState === "active" && (
            <p className="text-sm text-muted-foreground">{formatDuration(callDuration)}</p>
          )}
        </div>
        <Button variant="ghost" size="icon-sm" onClick={onClose}>
          <IconX className="size-4" />
        </Button>
      </div>

      {/* Phone Number Display */}
      <div className="p-4 bg-background">
        <div className="flex items-center gap-2 mb-2">
          <Input
            value={phoneNumber}
            onChange={(e) => setPhoneNumber(e.target.value)}
            placeholder="Enter phone number"
            className="text-2xl font-mono text-center h-14"
            disabled={callState === "active" || callState === "calling"}
          />
          {phoneNumber && callState === "idle" && (
            <Button variant="ghost" size="icon" onClick={handleBackspace}>
              <IconBackspace className="size-5" />
            </Button>
          )}
        </div>

        {/* Merged Calls Indicator */}
        {mergedCalls.length > 0 && (
          <div className="flex flex-wrap gap-2 mt-2">
            {mergedCalls.map((number, index) => (
              <div
                key={index}
                className="px-2 py-1 bg-blue-100 dark:bg-blue-900/30 rounded-md text-xs flex items-center gap-1"
              >
                <IconPhone className="size-3" />
                <span>{number}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Call Controls (Active Call) */}
      {(callState === "active" || callState === "calling") && (
        <div className="p-4 border-t border-b bg-muted/30">
          <div className="grid grid-cols-4 gap-3">
            <div className="flex flex-col items-center">
              <Button
                variant={isSpeakerOn ? "default" : "outline"}
                size="icon"
                className="rounded-full size-14"
                onClick={() => setIsSpeakerOn(!isSpeakerOn)}
              >
                <IconVolume className="size-5" />
              </Button>
              <span className="text-xs mt-1">Speaker</span>
            </div>

            <div className="flex flex-col items-center">
              <Button
                variant={isMuted ? "destructive" : "outline"}
                size="icon"
                className="rounded-full size-14"
                onClick={() => setIsMuted(!isMuted)}
              >
                {isMuted ? (
                  <IconMicrophoneOff className="size-5" />
                ) : (
                  <IconMicrophone className="size-5" />
                )}
              </Button>
              <span className="text-xs mt-1">Mute</span>
            </div>

            <div className="flex flex-col items-center">
              <Button
                variant="outline"
                size="icon"
                className="rounded-full size-14"
                onClick={handleMergeCall}
                disabled={!phoneNumber}
              >
                <IconPlus className="size-5" />
              </Button>
              <span className="text-xs mt-1">Merge</span>
            </div>

            <div className="flex flex-col items-center">
              <Button
                variant="ghost"
                size="icon"
                className="rounded-full size-14"
                onClick={() => setShowDialpad(!showDialpad)}
              >
                {showDialpad ? (
                  <IconChevronUp className="size-5" />
                ) : (
                  <IconChevronDown className="size-5" />
                )}
              </Button>
              <span className="text-xs mt-1">Keypad</span>
            </div>
          </div>
        </div>
      )}

      {/* Dialpad */}
      {showDialpad && callState !== "calling" && (
        <div className="p-4">
          <div className="grid grid-cols-3 gap-3">
            {dialpadKeys.flat().map((key) => (
              <Button
                key={key}
                variant="outline"
                size="lg"
                className={cn(
                  "h-16 text-2xl font-semibold rounded-full hover:bg-primary hover:text-primary-foreground transition-colors",
                  callState === "active" && "h-12 text-xl"
                )}
                onClick={() => handleDialpadClick(key)}
                disabled={callState === "active"}
              >
                {key}
              </Button>
            ))}
          </div>
        </div>
      )}

      {/* Call Actions */}
      <div className="p-4 flex gap-3 justify-center">
        {callState === "idle" && (
          <Button
            size="lg"
            className="rounded-full h-14 px-8 bg-green-600 hover:bg-green-700"
            onClick={handleCall}
            disabled={phoneNumber.length === 0}
          >
            <IconPhone className="size-5 mr-2" />
            Call
          </Button>
        )}

        {(callState === "active" || callState === "calling") && (
          <Button
            size="lg"
            className="rounded-full h-14 px-8 bg-red-600 hover:bg-red-700"
            onClick={handleEndCall}
          >
            <IconPhoneOff className="size-5 mr-2" />
            End Call
          </Button>
        )}
      </div>
    </Card>
  );
}
