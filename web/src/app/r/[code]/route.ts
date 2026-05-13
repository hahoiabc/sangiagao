import { NextResponse, type NextRequest } from "next/server";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/+$/, "") ||
  "http://localhost:8080/api/v1";

// Route handler /r/{code}: validate code with backend, set 30-day cookie,
// redirect to /dang-ky?ref={code}. Lives at /r/[code]/route.ts because
// Server Components can't set cookies in Next.js 16 — only Route Handlers can.
export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ code: string }> },
) {
  const { code: rawCode } = await params;
  const code = (rawCode || "").toUpperCase().replace(/[^A-Z0-9]/g, "");

  let valid = false;
  if (code.length >= 4 && code.length <= 8) {
    try {
      const r = await fetch(`${API_BASE}/referral/resolve/${code}`, {
        cache: "no-store",
      });
      valid = r.ok;
    } catch {
      valid = false;
    }
  }

  // Build absolute URL from forwarded headers (nginx upstream). req.url uses
  // the internal container URL (0.0.0.0:3001) which is not what we want to
  // redirect the browser to.
  const host =
    req.headers.get("x-forwarded-host") ||
    req.headers.get("host") ||
    "sangiagao.vn";
  const proto = req.headers.get("x-forwarded-proto") || "https";
  const path = valid ? `/dang-ky?ref=${code}` : "/dang-ky";
  const res = NextResponse.redirect(`${proto}://${host}${path}`);
  if (valid) {
    res.cookies.set("ref_code", code, {
      maxAge: 60 * 60 * 24 * 30, // 30 days
      path: "/",
      sameSite: "lax",
      httpOnly: false, // signup form reads this client-side as fallback
    });
  }
  return res;
}
