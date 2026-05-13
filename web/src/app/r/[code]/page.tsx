import { cookies } from "next/headers";
import { redirect } from "next/navigation";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/+$/, "") ||
  "http://localhost:8080/api/v1";

type Props = { params: Promise<{ code: string }> };

// Landing route /r/{code}: validate code with backend, set 30d cookie,
// redirect to /dang-ky?ref={code}. SSR — runs on server so cookie is set
// via Next.js cookies API (httpOnly off so /dang-ky JS can read fallback).
export default async function ReferralLandingPage({ params }: Props) {
  const { code: rawCode } = await params;
  const code = (rawCode || "").toUpperCase().replace(/[^A-Z0-9]/g, "");

  // Validate against backend; on any failure or unknown code, redirect to
  // /dang-ky without cookie so signup still works.
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

  if (valid) {
    const jar = await cookies();
    jar.set("ref_code", code, {
      maxAge: 60 * 60 * 24 * 30, // 30 days
      path: "/",
      sameSite: "lax",
      httpOnly: false, // signup form needs to read this client-side
    });
    redirect(`/dang-ky?ref=${code}`);
  }
  redirect("/dang-ky");
}
