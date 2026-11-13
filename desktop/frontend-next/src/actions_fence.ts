import type { UIAction } from "./actions";

export const FENCE_ACTIONS: UIAction[] = [
  { id: "code.download", kind: "interaction", target: "none", proof: "interaction" },
  { id: "code.fold", kind: "view", target: "none", proof: "interaction" },
  { id: "diagram.preview", kind: "view", target: "none", proof: "interaction" },
  { id: "diagram.source", kind: "view", target: "none", proof: "interaction" },
  { id: "diagram.zoom", kind: "view", target: "none", proof: "interaction" },
  { id: "diagram.zoom-reset", kind: "view", target: "none", proof: "interaction" },
  { id: "diagram.pan", kind: "view", target: "none", proof: "interaction" },
  { id: "diagram.fullscreen", kind: "view", target: "none", proof: "interaction" },
];
