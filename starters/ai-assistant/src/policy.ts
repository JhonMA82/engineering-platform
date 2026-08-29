export type ToolName = "search" | "read-document" | "send-message";

const roleTools: Record<string, readonly ToolName[]> = {
  reader: ["search", "read-document"],
  operator: ["search", "read-document", "send-message"],
};

export function canUseTool(role: string, tool: ToolName): boolean {
  return roleTools[role]?.includes(tool) ?? false;
}
