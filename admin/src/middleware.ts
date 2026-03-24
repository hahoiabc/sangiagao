import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const nonce = Buffer.from(crypto.randomUUID()).toString("base64");

  const isProd = process.env.NODE_ENV === "production";

  const csp = isProd
    ? [
        "default-src 'self'",
        `script-src 'self' 'nonce-${nonce}' 'strict-dynamic'`,
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' https://sangiagao.vn data: blob:",
        "font-src 'self' data:",
        "connect-src 'self' https://sangiagao.vn wss://sangiagao.vn",
        "frame-ancestors 'none'",
      ].join("; ")
    : [
        "default-src 'self'",
        `script-src 'self' 'nonce-${nonce}' 'strict-dynamic' 'unsafe-eval'`,
        "style-src 'self' 'unsafe-inline'",
        "img-src 'self' http: https: data: blob:",
        "font-src 'self' data:",
        "connect-src 'self' http://localhost:* ws://localhost:*",
        "frame-ancestors 'none'",
      ].join("; ");

  const requestHeaders = new Headers(request.headers);
  requestHeaders.set("x-nonce", nonce);

  const response = NextResponse.next({ request: { headers: requestHeaders } });
  response.headers.set("Content-Security-Policy", csp);

  return response;
}

export const config = {
  matcher: [
    { source: "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)" },
  ],
};
