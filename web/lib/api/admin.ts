import { apiFetch } from "./client";
import type { Report, PaginatedResponse } from "@/lib/types";

export async function getReports(params?: {
  status?: string;
  cursor?: string;
  limit?: number;
}): Promise<PaginatedResponse<Report>> {
  const search = new URLSearchParams();
  if (params?.status) search.set("status", params.status);
  if (params?.cursor) search.set("cursor", params.cursor);
  if (params?.limit) search.set("limit", params.limit.toString());
  const qs = search.toString();
  return apiFetch(`/admin/reports${qs ? `?${qs}` : ""}`);
}

export async function updateReport(
  reportId: string,
  status: string
): Promise<{ report: Report }> {
  return apiFetch(`/admin/reports/${reportId}`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}
