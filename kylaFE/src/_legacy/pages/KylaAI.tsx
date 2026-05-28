import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { IconSend, IconRobot, IconSparkles } from "@tabler/icons-react";
import PageContainer from "@/components/page-container";

const examplePrompts = [
  "How do I create a new support ticket?",
  "Show me today's pending tasks",
  "What are the top leads this week?",
  "Generate a customer satisfaction report",
  "Help me write an email response",
  "Summarize recent customer feedback",
];

type Message = {
  id: number;
  role: "user" | "assistant";
  content: string;
  timestamp: Date;
};

export default function KylaAIPage() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 1,
      role: "assistant",
      content:
        "Hello! I'm Kyla AI, your intelligent assistant. I can help you with support tasks, data analysis, report generation, and much more. How can I assist you today?",
      timestamp: new Date(),
    },
  ]);
  const [input, setInput] = useState("");

  const handleSend = () => {
    if (!input.trim()) return;

    const userMessage: Message = {
      id: messages.length + 1,
      role: "user",
      content: input,
      timestamp: new Date(),
    };

    setMessages([...messages, userMessage]);
    setInput("");

    // Simulate AI response
    setTimeout(() => {
      const aiMessage: Message = {
        id: messages.length + 2,
        role: "assistant",
        content:
          "I understand your request. This is a demo response. In a production environment, I would process your query and provide helpful insights based on your system data.",
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, aiMessage]);
    }, 1000);
  };

  const handleExampleClick = (prompt: string) => {
    setInput(prompt);
  };

  return (
    <PageContainer>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-3xl font-bold">
            <IconRobot className="size-8" />
            Kyla AI
          </h1>
          <p className="text-muted-foreground">
            Your intelligent copilot for support and operations
          </p>
        </div>
        <Badge variant="secondary" className="gap-1">
          <IconSparkles className="size-3" />
          AI Powered
        </Badge>
      </div>

      {/* Example Prompts */}
      {messages.length === 1 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Example Questions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-2 md:grid-cols-2">
              {examplePrompts.map((prompt, index) => (
                <Button
                  key={index}
                  variant="outline"
                  className="h-auto justify-start whitespace-normal p-3 text-left"
                  onClick={() => handleExampleClick(prompt)}
                >
                  <IconSparkles className="mr-2 size-4 shrink-0" />
                  {prompt}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Chat Messages */}
      <Card className="flex-1 overflow-hidden">
        <CardContent className="flex h-full flex-col p-4">
          <div className="flex-1 space-y-4 overflow-y-auto">
            {messages.map((message) => (
              <div
                key={message.id}
                className={`flex gap-3 ${
                  message.role === "user" ? "justify-end" : "justify-start"
                }`}
              >
                {message.role === "assistant" && (
                  <div className="bg-primary flex size-8 shrink-0 items-center justify-center rounded-full">
                    <IconRobot className="text-primary-foreground size-5" />
                  </div>
                )}
                <div
                  className={`max-w-[70%] rounded-lg px-4 py-2 ${
                    message.role === "user"
                      ? "bg-primary text-primary-foreground"
                      : "bg-muted"
                  }`}
                >
                  <p className="text-sm">{message.content}</p>
                  <p
                    className={`mt-1 text-xs ${
                      message.role === "user"
                        ? "text-primary-foreground/70"
                        : "text-muted-foreground"
                    }`}
                  >
                    {message.timestamp.toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </p>
                </div>
                {message.role === "user" && (
                  <div className="bg-secondary flex size-8 shrink-0 items-center justify-center rounded-full">
                    <span className="text-secondary-foreground text-sm font-semibold">
                      You
                    </span>
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Input Area */}
          <div className="mt-4 flex gap-2">
            <Input
              placeholder="Ask Kyla AI anything..."
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  handleSend();
                }
              }}
              className="flex-1"
            />
            <Button onClick={handleSend} className="gap-2">
              <IconSend className="size-4" />
              Send
            </Button>
          </div>
        </CardContent>
      </Card>
    </PageContainer>
  );
}
