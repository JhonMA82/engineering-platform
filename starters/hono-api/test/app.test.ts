import assert from "node:assert/strict";
import test from "node:test";
import { app } from "../src/app.js";

test("health contract", async () => {
  const response = await app.request("/health");
  assert.equal(response.status, 200);
  assert.deepEqual(await response.json(), { status: "ok" });
});

test("rejects an invalid request", async () => {
  const response = await app.request("/api/requests", { method: "POST", body: "{}", headers: { "content-type": "application/json" } });
  assert.equal(response.status, 400);
});
