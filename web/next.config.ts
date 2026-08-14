import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // The repo root has its own package.json (a task runner for `pnpm dev`) and
  // therefore its own lockfile. Next sees two lockfiles, infers the repo root
  // as the workspace root, and warns. This app is self-contained in web/, so
  // pin the root explicitly rather than let it guess.
  turbopack: {
    root: __dirname,
  },
};

export default nextConfig;
