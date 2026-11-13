import { useEffect, useMemo, useRef, useState } from "react";
import { t } from "../i18n";
import { CopyButton } from "./CopyButton";
import { ImageViewer } from "./ImagePreview";
import { listenAction } from "./listen";
import { StudioIcon } from "./StudioIcon";
import { useZoomPan, wheelFactor, type ZoomPan } from "./zoompan";
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

/** The chrome every rendered fence shares: Preview/Source, zoom, full screen, copy, failure, and the picture.
 *  A live fence stays source; a failed one shows its source beside the reason. */
export function DiagramFrame({ kind, label, copyLabel, failedLabel, pendingLabel, source, live, picture }: FrameProps) {
  const [mode, setMode] = useState<"preview" | "source">("preview");
  const [full, setFull] = useState(false);
  const current = live ? null : picture;
  const showSource = live || mode === "source" || !!current?.error;
  const svg = showSource ? undefined : current?.svg;
  const src = useMemo(() => (svg ? `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}` : ""), [svg]);
  const view = useZoomPan(src);
  const choose = (next: "preview" | "source") => {
    view.reset();
    setMode(next);
  };
  return (
    <section className="diagram-wrap" data-kind={kind} aria-label={label}>
      <div className="code-hd">
        <span className="code-lang">{kind}</span>
        <span className="code-acts">
          {src && (
            <>
              <button type="button" className="code-act" data-action="diagram.zoom" data-value="out" aria-label={t("chat.zoom_out")} title={t("chat.zoom_out")} onClick={() => view.zoom(0.8)}><StudioIcon name="zoomOut" /></button>
              <button type="button" className="diagram-mode diagram-level" data-action="diagram.zoom-reset" aria-label={t("恢复原始大小")} title={t("恢复原始大小")} onClick={view.reset}>{Math.round(view.scale * 100)}%</button>
              <button type="button" className="code-act" data-action="diagram.zoom" data-value="in" aria-label={t("chat.zoom_in")} title={t("chat.zoom_in")} onClick={() => view.zoom(1.25)}><StudioIcon name="zoomIn" /></button>
              <button type="button" className="code-act" data-action="diagram.fullscreen" aria-label={t("全屏查看")} title={t("全屏查看")} onClick={() => setFull(true)}><StudioIcon name="expand" /></button>
            </>
          )}
          <button type="button" className="diagram-mode" data-action="diagram.preview" aria-pressed={!showSource} onClick={() => choose("preview")}>{t("预览")}</button>
          <button type="button" className="diagram-mode" data-action="diagram.source" aria-pressed={!!showSource} onClick={() => choose("source")}>{t("源代码")}</button>
          {!live && <CopyButton text={source} iconOnly className="code-act" label={copyLabel} />}
        </span>
      </div>
      {current?.error && <p className="diagram-error" role="alert">{failedLabel}: {current.error}</p>}
      {showSource ? <pre className="diagram-source"><code>{source}</code></pre> : src ? (
        <DiagramPicture src={src} label={label} view={view} />
      ) : <p className="diagram-loading" role="status">{pendingLabel}</p>}
      {full && src && <ImageViewer src={src} name={label} download={`${kind}.svg`} onClose={() => setFull(false)} />}
    </section>
  );
}

/** Only ever an <img>: an SVG document inlined into the page is live markup that can script the host.
 *  Ctrl + wheel (and a trackpad pinch, which arrives as one) zooms here; a plain wheel scrolls the chat. */
function DiagramPicture({ src, label, view }: { src: string; label: string; view: ZoomPan }) {
  const box = useRef<HTMLDivElement>(null);
  const img = useRef<HTMLImageElement>(null);
  const { zoom } = view;
  useEffect(() => {
    const el = box.current;
    if (!el) return;
    const onWheel = (e: WheelEvent) => {
      if (!e.ctrlKey) return;
      e.preventDefault();
      const rect = el.getBoundingClientRect();
      const pic = img.current;
      const cx = pic ? pic.offsetLeft + pic.offsetWidth / 2 : rect.width / 2;
      const cy = pic ? pic.offsetTop + pic.offsetHeight / 2 : rect.height / 2;
      zoom(wheelFactor(e), { x: e.clientX - rect.left - cx, y: e.clientY - rect.top - cy });
    };
    // React registers wheel listeners as passive, and a passive listener cannot keep the chat from scrolling.
    return listenAction(el, "wheel", { action: "diagram.zoom", listener: onWheel as EventListener }, { passive: false });
  }, [zoom]);
  return (
    <div
      ref={box}
      className="diagram-preview"
      title={t("Ctrl + 滚轮缩放 · 拖动平移 · 双击复原")}
      data-action-pointerdown="diagram.pan"
      data-action-pointermove="diagram.pan"
      data-action-pointerup="diagram.pan"
      data-action-pointercancel="diagram.pan"
      data-action-dblclick="diagram.zoom-reset"
      onPointerDown={(ev) => {
        if (ev.target === img.current && ev.pointerType !== "touch") view.grab(ev);
      }}
      onPointerMove={view.move}
      onPointerUp={view.release}
      onPointerCancel={view.release}
      onDoubleClick={view.reset}
    >
      <img ref={img} alt={label} src={src} draggable={false} style={{ transform: view.transform }} />
    </div>
  );
}
