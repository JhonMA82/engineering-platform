import { Hono } from "hono";
import { z } from "zod";

const requestSchema = z.object({ title: z.string().min(1), description: z.string().min(1) });

export const app = new Hono()
  .get("/health", (context) => context.json({ status: "ok" }))
  .post("/api/requests", async (context) => {
    const parsed = requestSchema.safeParse(await context.req.json());
    if (!parsed.success) return context.json({ error: "invalid-request" }, 400);
    return context.json({ id: crypto.randomUUID(), ...parsed.data, status: "received" }, 201);
  });
