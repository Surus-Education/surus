"use client";

import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useUIStore } from "@/lib/stores/uiStore";

export function AuthPrompt() {
  const { authPromptOpen, authPromptAction, closeAuthPrompt } = useUIStore();
  const apiBase = process.env.NEXT_PUBLIC_API_URL?.replace("/v1", "") || "http://localhost:8080";

  return (
    <Dialog open={authPromptOpen} onOpenChange={(open) => !open && closeAuthPrompt()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Sign in required</DialogTitle>
          <DialogDescription>
            Sign in to {authPromptAction || "continue"}.
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-3 pt-2">
          <Button asChild>
            <a href={`${apiBase}/v1/auth/google/start`}>Sign in with Google</a>
          </Button>
          <Button variant="outline" asChild>
            <a href="/signin">Sign in with email</a>
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
