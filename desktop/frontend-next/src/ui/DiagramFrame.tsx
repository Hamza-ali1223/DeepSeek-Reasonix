import { useMemo, useState } from "react";
import { t } from "../i18n";
import { CopyButton } from "./CopyButton";
import "./diagram.css";

/** A settled render: an already-sanitized SVG document, or why there is none. */
export type Picture = { svg: string; error?: never } | { error: string; svg?: never };

type FrameProps = {
  kind: string;
  label: string;
  copyLabel: string;
  failedLabel: string;
  pendingLabel?: string;
  source: string;
  live?: boolean;
  picture: Picture | null;
};

/** The chrome every rendered fence shares: Preview/Source, copy, failure, and the picture itself.
 *  A live fence stays source; a failed one shows its source beside the reason. */
export function DiagramFrame({ kind, label, copyLabel, failedLabel, pendingLabel, source, live, picture }: FrameProps) {
  const [mode, setMode] = useState<"preview" | "source">("preview");
  const current = live ? null : picture;
  const showSource = live || mode === "source" || !!current?.error;
  return (
    <section className="diagram-wrap" data-kind={kind} aria-label={label}>
      <div className="code-hd">
        <span className="code-lang">{kind}</span>
        <span className="code-acts">
          <button type="button" className="diagram-mode" data-action="diagram.preview" aria-pressed={!showSource} onClick={() => setMode("preview")}>{t("预览")}</button>
          <button type="button" className="diagram-mode" data-action="diagram.source" aria-pressed={!!showSource} onClick={() => setMode("source")}>{t("源代码")}</button>
          {!live && <CopyButton text={source} iconOnly className="code-act" label={copyLabel} />}
        </span>
      </div>
      {current?.error && <p className="diagram-error" role="alert">{failedLabel}: {current.error}</p>}
      {showSource ? <pre className="diagram-source"><code>{source}</code></pre> : current?.svg ? (
        <DiagramPicture svg={current.svg} label={label} />
      ) : <p className="diagram-loading" role="status">{pendingLabel}</p>}
    </section>
  );
}

/** Only ever an <img>: an SVG document inlined into the page is live markup that can script the host. */
function DiagramPicture({ svg, label }: { svg: string; label: string }) {
  const src = useMemo(() => `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`, [svg]);
  return <div className="diagram-preview"><img alt={label} src={src} /></div>;
}
