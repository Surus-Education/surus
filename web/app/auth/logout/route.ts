import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

// Must match the path REFRESH_COOKIE_PATH used when the cookie was set in
// app/auth/callback/route.ts (and api/internal/auth/service.go's
// refreshCookiePath) — deleting a cookie only overwrites the original if the
// path matches exactly, otherwise the browser just writes a second cookie.
const REFRESH_COOKIE_PATH = "/v1/auth/refresh";

// GET, not POST: this must be reachable via a top-level navigation (an <a>
// or window.location assignment), not a fetch. A fetch response can't set
// cookies visible to the browser's own jar for this origin the way a
// navigation's Set-Cookie can, and the whole point of this route is to
// clear the web-origin cookie jar that app/auth/callback/route.ts wrote.
export async function GET(request: NextRequest) {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("access_token")?.value;

  // Best-effort: tell the API to revoke the refresh token / handoff codes
  // server-side. Forward the web-origin access_token as a Cookie header,
  // same pattern as lib/api/server/*.ts's serverFetch. If this fails or the
  // API is unreachable, still fall through and clear the local jar below —
  // a user who clicks sign out must end up signed out locally regardless.
  if (accessToken) {
    try {
      await fetch(`${API_URL}/auth/logout`, {
        method: "POST",
        headers: { Cookie: `access_token=${accessToken}` },
        cache: "no-store",
      });
    } catch {
      // ignore — clear local cookies regardless
    }
  }

  cookieStore.delete({ name: "access_token", path: "/" });
  cookieStore.delete({ name: "refresh_token", path: REFRESH_COOKIE_PATH });

  return NextResponse.redirect(new URL("/", request.url));
}
