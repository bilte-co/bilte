#!/bin/sh

# Bring up Tailscale in the background so a slow or failed login never delays
# the web server. Fly's readiness check expects :8080 shortly after boot.
if [ -n "${TAILSCALE_AUTHKEY}" ]; then
  (
    /app/tailscaled --state=/var/lib/tailscale/tailscaled.state --socket=/var/run/tailscale/tailscaled.sock &
    /app/tailscale up --auth-key="${TAILSCALE_AUTHKEY}" --hostname=fly-bilte-web --timeout=30s \
      || echo "tailscale up failed; continuing without tailscale" >&2
  ) &
else
  echo "TAILSCALE_AUTHKEY not set; skipping tailscale" >&2
fi

exec /app/bilte web
