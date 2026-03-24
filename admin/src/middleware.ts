import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const isProd = process.env.NODE_ENV === "production";

  const csp = isProd
    ? [
        "default-src 'self'",
        "script-src 'self' 'unsafe-inline'",
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' https://sangiagao.vn https://admin.sangiagao.vn data: blob:",
        "font-src 'self' data:",
        "connect-src 'self' https://admin.sangiagao.vn https://sangiagao.vn",
        "frame-ancestors 'none'",
      ].join("; ")
    : [
        "default-src 'self'",
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' http: https: data: blob:",
        "font-src 'self' data:",
        "connect-src 'self' http://localhost:* ws://localhost:*",
        "frame-ancestors 'none'",
      ].join("; ");

  const response = NextResponse.next();
  response.headers.set("Content-Security-Policy", csp);

  return response;
}

export const config = {
  matcher: [
    { source: "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)" },
  ],
};
