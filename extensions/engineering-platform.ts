import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import { basename, dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const PACKAGE_ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const ENG = join(PACKAGE_ROOT, "eng");
const PACKAGE_VERSION = JSON.parse(readFileSync(join(PACKAGE_ROOT, "package.json"), "utf8")).version as string;

function projectFiles(cwd: string): string[] {
  return readdirSync(cwd).filter((name) => ![".git", ".atl", ".gitignore"].includes(name));
}

export default function engineeringPlatform(pi: ExtensionAPI) {
  pi.registerCommand("new-project", {
    description: "Definir y preparar un proyecto nuevo con Engineering Platform",
    handler: async (_args, ctx) => {
      if (!ctx.hasUI) return;
      if (existsSync(join(ctx.cwd, ".engineering", "project.json"))) {
        ctx.ui.notify("Este directorio ya contiene un proyecto. Usa /engineering-status.", "warning");
        return;
      }
      const folderName = basename(ctx.cwd);
      if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(folderName)) {
        ctx.ui.notify(
          "La carpeta debe usar kebab-case. Usa: eng start <nombre-kebab-case>",
          "error",
        );
        return;
      }
      const files = projectFiles(ctx.cwd);
      if (files.length > 0) {
        ctx.ui.notify(
          "El directorio no está vacío. Para evitar proyectos anidados ejecuta en otra terminal: eng start <nombre>",
          "error",
        );
        return;
      }
      pi.setSessionName(`new:${folderName}`);
      pi.appendEntry("engineering-platform:discovery", {
        version: PACKAGE_VERSION,
        cwd: ctx.cwd,
        status: "started",
      });
      pi.sendUserMessage(
        `/skill:project-discovery El proyecto se creará en ${JSON.stringify(ctx.cwd)}. ` +
          `Usa ${JSON.stringify(ENG)} como ejecutable de Engineering Platform.`,
        { expandPromptTemplates: true },
      );
    },
  });

  pi.registerCommand("engineering-status", {
    description: "Comprobar el manifest del proyecto actual",
    handler: async (_args, ctx) => {
      const project = join(ctx.cwd, ".engineering", "project.json");
      if (!existsSync(project)) {
        ctx.ui.notify("No hay manifest. Usa /new-project dentro de una carpeta vacía.", "warning");
        return;
      }
      const result = await pi.exec("python3", [ENG, "doctor", "--project", ctx.cwd], {
        cwd: ctx.cwd,
        timeout: 15_000,
      });
      ctx.ui.notify(
        result.code === 0 ? "Engineering Platform: proyecto coherente." : "Engineering Platform detectó errores.",
        result.code === 0 ? "info" : "error",
      );
      pi.sendMessage(
        {
          customType: "engineering-platform:doctor",
          content: result.stdout || result.stderr,
          display: true,
        },
        { triggerTurn: false },
      );
    },
  });

  pi.registerCommand("evolve-project", {
    description: "Agregar una aplicación o capacidad sin regenerar el proyecto",
    handler: async (_args, ctx) => {
      const project = join(ctx.cwd, ".engineering", "project.json");
      if (!existsSync(project)) {
        ctx.ui.notify("No hay un proyecto Engineering en esta carpeta.", "error");
        return;
      }
      pi.setSessionName(`evolve:${basename(ctx.cwd)}`);
      pi.appendEntry("engineering-platform:evolution", {
        version: PACKAGE_VERSION,
        cwd: ctx.cwd,
        status: "started",
      });
      pi.sendUserMessage(
        `/skill:project-evolution El proyecto está en ${JSON.stringify(ctx.cwd)}. ` +
          `Usa ${JSON.stringify(ENG)} como ejecutable de Engineering Platform.`,
        { expandPromptTemplates: true },
      );
    },
  });
}
