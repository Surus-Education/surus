"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useAuth } from "@/hooks/useAuth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { updateMe, deleteMe, exportData } from "@/lib/api/users";
import { toast } from "sonner";

export default function SettingsPage() {
  const { user, logout } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const [displayName, setDisplayName] = useState(user?.display_name || "");
  const [bio, setBio] = useState(user?.bio || "");

  const updateMutation = useMutation({
    mutationFn: () => updateMe({ display_name: displayName, bio }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["auth", "me"] });
      toast.success("Profile updated");
    },
    onError: () => toast.error("Failed to update profile"),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteMe,
    onSuccess: async () => {
      setDeleteDialogOpen(false);
      toast.success("Account deletion scheduled. You have 30 days to change your mind.");
      await logout();
      router.push("/");
    },
    onError: () => toast.error("Failed to schedule deletion"),
  });

  const exportMutation = useMutation({
    mutationFn: exportData,
    onSuccess: () => toast.success("Export started. You'll receive an email with a download link."),
    onError: () => toast.error("Failed to start export"),
  });

  if (!user) return null;

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">Settings</h1>

      <section className="space-y-4 mb-8">
        <h2 className="text-lg font-semibold">Profile</h2>
        <div>
          <Label>Display name</Label>
          <Input value={displayName} onChange={(e) => setDisplayName(e.target.value)} maxLength={100} />
        </div>
        <div>
          <Label>Bio</Label>
          <Textarea value={bio} onChange={(e) => setBio(e.target.value)} maxLength={500} />
        </div>
        <Button onClick={() => updateMutation.mutate()} disabled={updateMutation.isPending}>
          {updateMutation.isPending ? "Saving..." : "Save changes"}
        </Button>
      </section>

      <Separator className="my-8" />

      <section className="space-y-4 mb-8">
        <h2 className="text-lg font-semibold">Data export</h2>
        <p className="text-sm text-muted-foreground">
          Export all your data including courses, progress, and quiz attempts.
        </p>
        <Button variant="outline" onClick={() => exportMutation.mutate()} disabled={exportMutation.isPending}>
          {exportMutation.isPending ? "Starting..." : "Export my data"}
        </Button>
      </section>

      <Separator className="my-8" />

      <section className="space-y-4">
        <h2 className="text-lg font-semibold text-destructive">Danger zone</h2>
        <p className="text-sm text-muted-foreground">
          Deleting your account will schedule it for permanent deletion after 30 days.
          You can sign back in during that period to cancel.
        </p>
        <Button variant="destructive" onClick={() => setDeleteDialogOpen(true)}>
          Delete account
        </Button>
      </section>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete your account?</DialogTitle>
            <DialogDescription>This action schedules your account for deletion.</DialogDescription>
          </DialogHeader>
          <div className="text-sm space-y-2">
            <p><strong>What will be deleted:</strong></p>
            <ul className="list-disc pl-4 space-y-1">
              <li>Your profile and personal data</li>
              <li>All private and unlisted courses</li>
              <li>Your saves, completions, and quiz attempts</li>
            </ul>
            <p><strong>What will remain:</strong></p>
            <ul className="list-disc pl-4 space-y-1">
              <li>Your public courses (attributed to &quot;Deleted user&quot;)</li>
            </ul>
            <p>You have <strong>30 days</strong> to sign back in and cancel.</p>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
            <Button
              variant="destructive"
              onClick={() => deleteMutation.mutate()}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? "Deleting..." : "Confirm delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
