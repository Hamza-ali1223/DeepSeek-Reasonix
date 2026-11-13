import {
  useEffect,
  useRef,
  useState,
  type MouseEvent,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import { t } from "../i18n";
import { useEscape } from "./dismiss";
import { StudioIcon } from "./StudioIcon";
import { useZoomPan } from "./zoompan";
import "./image-preview.css";

export function ImagePreview({
  src,
  name,
  className,
  children,
}: {
  src?: string;
  name: string;
  className?: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  useEffect(() => setOpen(false), [src]);
  return (
    <>
      <button
        type="button"
        className={["image-trigger", className].filter(Boolean).join(" ")}
        disabled={!src}
        data-action-click="image.open"
        data-action-mousedown="image.open"
        aria-label={t("chat.preview_image_name", { name })}
        title={t("chat.preview_image_name", { name })}
        onClick={() => setOpen(true)}
        onMouseDown={(ev) => {
          if (ev.detail > 1) ev.preventDefault();
        }}
      >
        {children}
      </button>
      {open && src && (
        <ImageViewer src={src} name={name} onClose={() => setOpen(false)} />
      )}
    </>
  );
}

/** The full-window picture viewer. `download` is the file name a saved copy takes; it defaults to `name`. */
export function ImageViewer({
  src,
  name,
  download: file,
  onClose,
}: {
  src: string;
  name: string;
  download?: string;
  onClose: () => void;
}) {
  const close = useRef<HTMLButtonElement>(null);
  const [failed, setFailed] = useState(false);
  const view = useZoomPan(src);
  const [size, setSize] = useState<[number, number] | null>(null);
  useEscape(true, onClose);
  const backdrop = (ev: MouseEvent<HTMLDivElement>) => {
    if (view.endedDrag()) return;
    if (ev.target === ev.currentTarget && ev.detail < 2) onClose();
  };
  const download = () => {
    const link = document.createElement("a");
    link.href = src;
    link.download = file ?? name;
    link.click();
  };
  const zoom = (factor: number) => view.zoom(factor);
  useEffect(() => {
    const previous = document.activeElement;
    const overflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    close.current?.focus();
    return () => {
      document.body.style.overflow = overflow;
      if (previous instanceof HTMLElement && previous.isConnected)
        previous.focus();
    };
  }, []);

  // Transcript and composer ancestors can scroll or carry backdrop filters;
  // the body portal keeps the viewer outside those containing blocks.
  return createPortal(
    <div
      className="image-preview"
      role="dialog"
      aria-modal="true"
      aria-label={name}
      data-action-click="image.close"
      data-action-keydown="image.focus-trap"
      data-action-mousedown="image.focus-trap"
      data-action-pointerdown="image.pan"
      onClick={backdrop}
      onPointerDownCapture={view.press}
      onMouseDown={(ev) => {
        if (!(ev.target as HTMLElement).closest("button")) ev.preventDefault();
      }}
      onKeyDown={(ev) => {
        if (ev.key === "Tab") {
          ev.preventDefault();
          close.current?.focus();
        }
      }}
    >
      <div className="image-preview-bar">
        <span className="image-preview-name">{name}</span>
        {size && (
          <span className="image-preview-size">
            {size[0]} × {size[1]}
          </span>
        )}
        <span className="image-preview-gap" />
        <button
          type="button"
          data-action="image.zoom"
          data-value="in"
          aria-label={t("chat.zoom_in")}
          title={t("chat.zoom_in")}
          onClick={() => zoom(1.25)}
        >
          <StudioIcon name="zoomIn" />
        </button>
        <button
          type="button"
          data-action="image.zoom"
          data-value="out"
          aria-label={t("chat.zoom_out")}
          title={t("chat.zoom_out")}
          onClick={() => zoom(0.8)}
        >
          <StudioIcon name="zoomOut" />
        </button>
        <button
          type="button"
          data-action="image.download"
          aria-label={t("chat.image_download")}
          title={t("chat.image_download")}
          onClick={download}
        >
          <StudioIcon name="download" />
        </button>
        <button
          type="button"
          ref={close}
          className="image-preview-close"
          data-action="image.close"
          aria-label={t("chat.close_image_preview")}
          onClick={onClose}
        >
          <StudioIcon name="close" />
        </button>
      </div>
      <div
        className="image-preview-view"
        data-action-click="image.close"
        data-action-wheel="image.zoom"
        data-action-pointerdown="image.pan"
        data-action-pointermove="image.pan"
        data-action-pointerup="image.pan"
        onClick={backdrop}
        onWheel={(ev) => zoom(ev.deltaY < 0 ? 1.1 : 1 / 1.1)}
        onPointerDown={(ev) => {
          if (ev.target instanceof HTMLImageElement) view.grab(ev);
        }}
        onPointerMove={view.move}
        onPointerUp={view.release}
      >
        {failed ? (
          <p role="status">{t("chat.image_unavailable")}</p>
        ) : (
          <img
            src={src}
            alt={name}
            draggable={false}
            style={{ transform: view.transform }}
            onLoad={(ev) =>
              setSize([
                ev.currentTarget.naturalWidth,
                ev.currentTarget.naturalHeight,
              ])
            }
            onError={() => setFailed(true)}
          />
        )}
      </div>
      <div className="image-preview-hint">{t("chat.image_preview_hint")}</div>
    </div>,
    document.body,
  );
}
