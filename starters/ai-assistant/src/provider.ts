export type Citation = { sourceId: string; label: string };
export type AssistantAnswer = { text: string; citations: Citation[] };

export interface AssistantProvider {
  answer(input: { message: string; sourceIds: string[] }): Promise<AssistantAnswer>;
}

export const unconfiguredProvider: AssistantProvider = {
  async answer() { throw new Error("Configure an AssistantProvider before serving chat requests"); },
};
