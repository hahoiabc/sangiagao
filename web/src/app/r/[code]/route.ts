import { NextResponse, type NextRequest } from "next/server";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/+$/, "") ||
  "http://localhost:8080/api/v1";

// Route handler /r/{code}:
// 1. Validate code (must belong to an active aff user)
// 2. Set 30-day cookie for web register fallback
// 3. Detect OS via User-Agent → redirect:
//    - Android → Play Store with ?referrer={code} (Install Referrer API)
//    - iOS     → App Store + try Universal Link if app installed
//    - Other   → /cai-app?ref={code} landing page (manual choice)
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

  // Build absolute base URL from forwarded headers (nginx upstream sets these).
  const host =
    req.headers.get("x-forwarded-host") ||
    req.headers.get("host") ||
    "sangiagao.vn";
  const proto = req.headers.get("x-forwarded-proto") || "https";
  const base = `${proto}://${host}`;

  // Show landing page for all (no auto-redirect). User picks platform.
  // The landing page itself contains the OS-aware install buttons.
  const target = valid ? `${base}/cai-app?ref=${code}` : `${base}/cai-app`;
  const res = NextResponse.redirect(target);
  if (valid) {
    res.cookies.set("ref_code", code, {
      maxAge: 60 * 60 * 24 * 30, // 30 days
      path: "/",
      sameSite: "lax",
      httpOnly: false,
    });
  }
  return res;
}

