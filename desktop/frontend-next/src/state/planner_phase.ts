import { t } from "../i18n";
import type { WireEvent } from "../port/wire";
import { nextId } from "./ids";
import { sealSay } from "./say";
import type { SessionState } from "./session_types";

const LABEL: Record<string, string> = {
  "planner.requested": "session.planner_requested",
  "planner.executor_starting": "session.planner_executor_starting",
  "planner.consulting": "session.planner_consulting",
  "planner.executor_resuming": "session.planner_executor_resuming",
};

export function foldPlannerPhase(state: SessionState, event: WireEvent): SessionState {
  const wording = event.code ? LABEL[event.code] : undefined;
  if (!wording) return state;
  const text = t(wording, { model: event.modelRef ?? "" });
  const items = sealSay(state.items);
  return { ...state, items: [...items, { t: "notice", id: nextId(), level: "info", text, code: event.code }] };
}
