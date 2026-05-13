"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSearchParams } from "next/navigation";
import { getReports, updateReport } from "@/lib/api/admin";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { Suspense } from "react";

function ReportsContent() {
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const status = searchParams.get("status") || "open";

  const { data, isLoading } = useQuery({
    queryKey: ["admin", "reports", status],
    queryFn: () => getReports({ status }),
  });

  const updateMutation = useMutation({
    mutationFn: ({ reportId, status }: { reportId: string; status: string }) =>
      updateReport(reportId, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin", "reports"] });
      toast.success("Report updated");
    },
    onError: () => toast.error("Failed to update report"),
  });

  const reports = data?.data ?? [];

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">Reports</h1>

      <div className="flex gap-2 mb-6">
        {["open", "reviewed", "actioned", "dismissed"].map((s) => (
          <Badge
            key={s}
            variant={status === s ? "default" : "outline"}
            className="cursor-pointer"
            onClick={() => {
              const params = new URLSearchParams(searchParams);
              params.set("status", s);
              window.history.replaceState(null, "", `?${params.toString()}`);
            }}
          >
            {s}
          </Badge>
        ))}
      </div>

      {isLoading && <p className="text-muted-foreground">Loading...</p>}

      {reports.length === 0 && !isLoading && (
        <p className="text-muted-foreground">No {status} reports.</p>
      )}

      <div className="space-y-3">
        {reports.map((report) => (
          <div key={report.id} className="border rounded-lg p-4">
            <div className="flex items-start justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <Badge variant="secondary">{report.target_type}</Badge>
                  <Badge variant="outline">{report.category}</Badge>
                  <span className="text-xs text-muted-foreground">
                    {new Date(report.created_at).toLocaleDateString()}
                  </span>
                </div>
                <p className="text-sm">Target: {report.target_id}</p>
                {report.body && <p className="text-sm text-muted-foreground">{report.body}</p>}
              </div>
              <div className="flex gap-1 shrink-0">
                {status === "open" && (
                  <>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => updateMutation.mutate({ reportId: report.id, status: "reviewed" })}
                      disabled={updateMutation.isPending}
                    >
                      Reviewed
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => updateMutation.mutate({ reportId: report.id, status: "actioned" })}
                      disabled={updateMutation.isPending}
                    >
                      Action
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => updateMutation.mutate({ reportId: report.id, status: "dismissed" })}
                      disabled={updateMutation.isPending}
                    >
                      Dismiss
                    </Button>
                  </>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function AdminReportsPage() {
  return (
    <Suspense fallback={<div className="mx-auto max-w-6xl px-4 py-8">Loading...</div>}>
      <ReportsContent />
    </Suspense>
  );
}
