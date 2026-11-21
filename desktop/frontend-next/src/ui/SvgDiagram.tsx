import { useEffect, useMemo, useState } from "react";
import { t } from "../i18n";
import { reason } from "../i18n/kernel";
import { DiagramFrame, type Picture } from "./DiagramFrame";
import { DIAGRAM_SOURCE_LIMIT, sanitizeInertSvg } from "./inertSvg";

const REDUCE = "(prefers-reduced-motion: reduce)";

function useReducedMotion(): boolean {
  const [reduce, setReduce] = useState(() => typeof matchMedia === "function" && matchMedia(REDUCE).matches);
  useEffect(() => {
    if (typeof matchMedia !== "function") return;
    const query = matchMedia(REDUCE);
    const sync = () => setReduce(query.matches);
    sync();
    query.addEventListener("change", sync);
    return () => query.removeEventListener("change", sync);
  }, []);
  return reduce;
}

function pictureOf(source: string, motion: boolean): Picture {
  if (source.length > DIAGRAM_SOURCE_LIMIT) {
    return { error: t("chat.svg_source_limit", { n: DIAGRAM_SOURCE_LIMIT.toLocaleString("en-US") }) };
  }
  try {
    return { svg: sanitizeInertSvg(source, { motion, localReferences: true }) };
  } catch (error) {
    return { error: reason(error) };
  }
}

/** An agent-written SVG fence. A streaming tail is never sanitized: half a document is not a picture.
 *  It animates unless the system asks for reduced motion, and then shows the drawing still. */
export function SvgDiagram({ source, live }: { source: string; live?: boolean }) {
  const reduce = useReducedMotion();
  const picture = useMemo(() => (live ? null : pictureOf(source, !reduce)), [source, live, reduce]);
  return (
    <DiagramFrame
      kind="svg"
      label={t("chat.svg_image")}
      copyLabel={t("chat.svg_copy_source")}
      failedLabel={t("chat.svg_preview_failed")}
      source={source}
      live={live}
      picture={picture}
    />
  );
}
