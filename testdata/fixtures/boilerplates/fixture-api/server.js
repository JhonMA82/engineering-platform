// Fixture API: offline test foundation standing in for a curated backend.
export function handler(req) {
  return { status: 200, body: "fixture-api ok: " + req.url };
}
