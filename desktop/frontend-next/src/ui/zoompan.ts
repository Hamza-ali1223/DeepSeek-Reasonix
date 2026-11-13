import { useCallback, useRef, useState, type PointerEvent } from "react";

export const ZOOM_MIN = 0.2;
export const ZOOM_MAX = 8;
const DRAG_SLOP = 4;

type Point = { x: number; y: number };
type View = Point & { key: string; scale: number };

const clamp = (scale: number) => Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, Math.round(scale * 1000) / 1000));

/** How far one wheel event zooms. A trackpad pinch sends many small deltas and a
 *  mouse notch one large one, so the step follows the delta rather than counting events. */
export function wheelFactor(e: { deltaY: number; deltaMode: number }): number {
  const pixels = e.deltaMode === 1 ? e.deltaY * 16 : e.deltaMode === 2 ? e.deltaY * 400 : e.deltaY;
  return Math.exp(-Math.max(-100, Math.min(100, pixels)) / 500);
}

/** Scale and offset of a picture drawn at its fitted size, as one CSS transform.
 *  `key` names what is shown: a different key starts again at 100%. A zoom `at` a point,
 *  given from the centre of the box, keeps what is under that point where it is. */
export function useZoomPan(key = "") {
  const [stored, setStored] = useState<View>({ key, scale: 1, x: 0, y: 0 });
  const view = stored.key === key ? stored : { key, scale: 1, x: 0, y: 0 };
  const drag = useRef<{ id: number; x: number; y: number; from: Point } | null>(null);
  const dragged = useRef(false);
  const update = useCallback(
    (next: (v: View) => View) => setStored((was) => next(was.key === key ? was : { key, scale: 1, x: 0, y: 0 })),
    [key],
  );
  const zoom = useCallback(
    (factor: number, at?: Point) =>
      update((v) => {
        const scale = clamp(v.scale * factor);
        if (!at) return { ...v, scale };
        const k = scale / v.scale;
        return { key: v.key, scale, x: at.x - (at.x - v.x) * k, y: at.y - (at.y - v.y) * k };
      }),
    [update],
  );
  const reset = useCallback(() => {
    drag.current = null;
    setStored({ key, scale: 1, x: 0, y: 0 });
  }, [key]);
  /** A new press starts a new gesture: whatever ended the last drag no longer applies. */
  const press = () => {
    dragged.current = false;
  };
  const grab = (ev: PointerEvent<HTMLElement>) => {
    press();
    if (ev.button !== 0) return;
    drag.current = { id: ev.pointerId, x: ev.clientX, y: ev.clientY, from: { x: view.x, y: view.y } };
    ev.currentTarget.setPointerCapture(ev.pointerId);
  };
  const move = (ev: PointerEvent<HTMLElement>) => {
    const d = drag.current;
    if (!d || d.id !== ev.pointerId) return;
    if (Math.abs(ev.clientX - d.x) + Math.abs(ev.clientY - d.y) > DRAG_SLOP) dragged.current = true;
    update((v) => ({ ...v, x: d.from.x + ev.clientX - d.x, y: d.from.y + ev.clientY - d.y }));
  };
  const release = () => {
    drag.current = null;
  };
  /** Whether the click now arriving ends a drag rather than being one. Pointer capture
   *  retargets that click to the capturing box, so a backdrop reads it as its own; the
   *  answer holds for every listener that click bubbles through, and the next press clears it. */
  const endedDrag = () => dragged.current;
  return {
    scale: view.scale,
    transform: `translate(${view.x}px, ${view.y}px) scale(${view.scale})`,
    zoom,
    reset,
    press,
    grab,
    move,
    release,
    endedDrag,
  };
}

export type ZoomPan = ReturnType<typeof useZoomPan>;
