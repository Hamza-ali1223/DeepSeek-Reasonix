import { useEffect, useState } from "react";
import { t } from "../i18n";
import { reason } from "../i18n/kernel";
import { DiagramFrame, type Picture } from "./DiagramFrame";

type Rendered = { source: string; theme: string } & Picture;

/** A streaming fence stays source text until its block settles. */
export function MermaidDiagram({ source, live }: { source: string; live?: boolean }) {
  const [theme, setTheme] = useState<"light" | "dark">(() => document.documentElement.dataset.theme === "dark" ? "dark" : "light");
  const [rendered, setRendered] = useState<Rendered | null>(null);
  useEffect(() => {
    const watch = new MutationObserver(() => setTheme(document.documentElement.dataset.theme === "dark" ? "dark" : "light"));
    watch.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });
    return () => watch.disconnect();
  }, []);
  useEffect(() => {
    if (live) return;
    let alive = true;
    // Ordinary Markdown and unsettled tails never need the renderer module or iframe.
    void import("./mermaid").then((engine) => engine.renderMermaid(source, theme)).then(
      (svg) => { if (alive) setRendered({ source, theme, svg }); },
      (error: unknown) => { if (alive) setRendered({ source, theme, error: reason(error) }); },
    );
    return () => { alive = false; };
  }, [source, theme, live]);
  const current = rendered?.source === source && rendered.theme === theme ? rendered : null;
  return (
    <DiagramFrame
      kind="mermaid"
      label={t("chat.mermaid_diagram")}
      copyLabel={t("chat.mermaid_copy_source")}
      failedLabel={t("chat.mermaid_render_failed")}
      pendingLabel={t("chat.mermaid_rendering")}
      source={source}
      live={live}
      picture={current}
    />
  );
}
