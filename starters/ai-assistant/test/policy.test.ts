import assert from "node:assert/strict";
import test from "node:test";
import { canUseTool } from "../src/policy.js";

test("tools are denied by default", () => assert.equal(canUseTool("unknown", "send-message"), false));
test("reader cannot trigger side effects", () => assert.equal(canUseTool("reader", "send-message"), false));
