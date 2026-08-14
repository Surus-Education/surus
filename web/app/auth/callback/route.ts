import { NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

// refreshCookiePath must match the API's own refresh_token cookie path
// (api/internal/auth/service.go's refreshCookiePath). It's mirrored here only
// so the two jars stay consistent with each other, not because this web-side
// cookie is read by a route at that path.
const REFRESH_COOKIE_PATH = "/v1/auth/refresh";

// Deployed, NEXT_PUBLIC_API_URL is an https:// origin; local dev is plain
// HTTP. Same coupling as api/internal/auth/service.go's cookieAttrs(): only
// set Secure (and therefore SameSite=None) when the connection can actually
// be secure, since browsers reject SameSite=None without Secure.
function cookieAttrs(): { secure: boolean; sameSite: "none" | "lax" } {
  let isHttps = false;
  try {
    isHttps = new URL(API_URL).protocol === "https:";
  } catch {
    isHttps = false;
  }
  return isHttps ? { secure: true, sameSite: "none" } : { secure: false, sameSite: "lax" };
}

export async function GET(request: NextRequest) {
  const code = request.nextUrl.searchParams.get("code");

  if (!code) {
    return NextResponse.redirect(new URL("/?error=auth_failed", request.url));
  }

  let tokens: { access_token?: string; refresh_token?: string };
  try {
    const res = await fetch(`${API_URL}/auth/handoff`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code }),
      cache: "no-store",
    });

    if (!res.ok) {
      return NextResponse.redirect(new URL("/?error=auth_failed", request.url));
    }

    tokens = await res.json();
  } catch {
    return NextResponse.redirect(new URL("/?error=auth_failed", request.url));
  }

  if (!tokens.access_token || !tokens.refresh_token) {
    return NextResponse.redirect(new URL("/?error=auth_failed", request.url));
  }

  const { secure, sameSite } = cookieAttrs();
  const cookieStore = await cookies();

  cookieStore.set("access_token", tokens.access_token, {
    httpOnly: true,
    secure,
    sameSite,
    path: "/",
    maxAge: 900,
  });

  cookieStore.set("refresh_token", tokens.refresh_token, {
    httpOnly: true,
    secure,
    sameSite,
    path: REFRESH_COOKIE_PATH,
    maxAge: 2592000,
  });

  return NextResponse.redirect(new URL("/dashboard", request.url));
}
