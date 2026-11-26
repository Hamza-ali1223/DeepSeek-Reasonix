import type { UIAction } from "./actions";

// The planner collaboration settings. The three inputs edit a draft; the save is
// the only one that reaches the kernel, writing the user config later sessions read.
export const PLANNER_ACTIONS: UIAction[] = [
  { id: "planner.research-rounds", kind: "view", target: "none", proof: "interaction" },
  { id: "planner.consultation", kind: "view", target: "none", proof: "interaction" },
  { id: "planner.consultation-limit", kind: "view", target: "none", proof: "interaction" },
  { id: "planner.tools", kind: "view", target: "none", proof: "interaction" },
  { id: "planner.consultation-timeout", kind: "view", target: "none", proof: "interaction" },
  { id: "planner.discard", kind: "view", target: "none", proof: "interaction" },
  { id: "planner.save", kind: "kernel-mutation", target: "none", proof: "authority-effect" },
];
