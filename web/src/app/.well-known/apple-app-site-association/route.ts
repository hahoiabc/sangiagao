import { NextResponse } from "next/server";

// Apple Universal Links Site Association file. Apple iOS fetches this from
// /.well-known/apple-app-site-association when the app is installed (and
// periodically thereafter). Must be served:
//   - over HTTPS
//   - with Content-Type: application/json
//   - HTTP 200, no redirect
// Routes listed in 'paths' will open the app instead of Safari when clicked.

const AASA = {
  applinks: {
    apps: [],
    details: [
      {
        appID: "4398LD7T8U.com.sangiagao.riceMarketplace",
        paths: ["/r/*", "/cai-app", "/cai-app/*"],
      },
    ],
  },
};

export const dynamic = "force-static";

export async function GET() {
  return NextResponse.json(AASA, {
    headers: {
      "Content-Type": "application/json",
      "Cache-Control": "public, max-age=3600",
    },
  });
}
