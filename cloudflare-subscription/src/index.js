const UPSTREAM_ORIGIN = "https://main.undervineyard.com";
const UPSTREAM_HOST = new URL(UPSTREAM_ORIGIN).host;

addEventListener("fetch", event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  const incomingUrl = new URL(request.url);
  const upstreamUrl = new URL(incomingUrl.pathname + incomingUrl.search, UPSTREAM_ORIGIN);
  const isNoStoreRoute = incomingUrl.pathname.startsWith("/v1/") || incomingUrl.pathname === "/ws";
  const isWebSocket = request.headers.get("Upgrade")?.toLowerCase() === "websocket";

  const headers = buildUpstreamHeaders(request, incomingUrl);

  const init = {
    method: request.method,
    headers,
    redirect: "manual",
    body: request.method === "GET" || request.method === "HEAD" ? undefined : request.body,
  };

  try {
    const upstreamResponse = await fetch(new Request(upstreamUrl, init), {
      cf: isNoStoreRoute ? { cacheTtl: 0, cacheEverything: false } : undefined,
    });

    if (isWebSocket) {
      return upstreamResponse;
    }

    return buildClientResponse(upstreamResponse, isNoStoreRoute);
  } catch (error) {
    return new Response(JSON.stringify({
      error: "upstream fetch failed",
      upstream: UPSTREAM_HOST,
      detail: error instanceof Error ? error.message : String(error),
    }), {
      status: 502,
      headers: {
        "Content-Type": "application/json; charset=utf-8",
        "Cache-Control": "no-store",
      },
    });
  }
}

function buildUpstreamHeaders(request, incomingUrl) {
  const headers = new Headers(request.headers);

  for (const name of Array.from(headers.keys())) {
    const lowerName = name.toLowerCase();
    if (
      lowerName === "host" ||
      lowerName === "referer" ||
      lowerName.startsWith("cf-") ||
      lowerName.startsWith("x-forwarded-") ||
      lowerName === "x-real-ip"
    ) {
      headers.delete(name);
    }
  }

  headers.set("Host", UPSTREAM_HOST);
  headers.set("Origin", UPSTREAM_ORIGIN);
  headers.set("X-Forwarded-Host", incomingUrl.host);
  headers.set("X-Forwarded-Proto", incomingUrl.protocol.replace(":", ""));

  const clientIP = request.headers.get("CF-Connecting-IP");
  if (clientIP) {
    headers.set("X-Real-IP", clientIP);
    headers.set("X-Forwarded-For", clientIP);
  }

  return headers;
}

function buildClientResponse(response, isNoStoreRoute) {
  const headers = new Headers(response.headers);

  if (isNoStoreRoute) {
    headers.set("Cache-Control", "no-store");
    headers.set("Pragma", "no-cache");
    headers.delete("ETag");
  }

  headers.delete("Content-Length");

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}
