import { Hono } from "hono";
import { z } from "zod";
import type { AssistantProvider } from "./provider.js";
import { unconfiguredProvider } from "./provider.js";

const requestSchema = z.object({ message: z.string().min(1), sourceIds: z.array(z.string()).default([]) });

export function createApp(provider: AssistantProvider = unconfiguredProvider) {
  return new Hono()
    .get("/health", (context) => context.json({ status: "ok" }))
    .post("/api/chat", async (context) => {
      const parsed = requestSchema.safeParse(await context.req.json());
      if (!parsed.success) return context.json({ error: "invalid-request" }, 400);
      const answer = await provider.answer(parsed.data);
      return context.json({ ...answer, audit: { action: "assistant.answer", sourceIds: parsed.data.sourceIds } });
    });
}
