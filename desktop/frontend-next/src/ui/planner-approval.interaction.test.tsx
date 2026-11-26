// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import "./testkit";
import { Pane } from "./Pane";
import { MockPort } from "../port/mock";
import type { RuntimeView } from "../port/hub";
import type { AgentPort } from "../port/port";
import type { WireEvent } from "../port/wire";

afterEach(() => { cleanup(); localStorage.clear(); vi.restoreAllMocks(); });

it("carries an accepted planner approval from the session stream to the next submission", async () => {
  const port = new MockPort();
  const submit = vi.spyOn(port as AgentPort, "submit").mockResolvedValue(undefined);
  const listeners = new Set<(event: WireEvent) => void>();
  vi.spyOn(port, "subscribe").mockImplementation((listener) => {
    listeners.add(listener);
    return () => listeners.delete(listener);
  });
  const { container } = render(<Pane port={port} rt={{ id: "planner", root: "/w", name: "work" } as RuntimeView}
    title="work" active visible sideHost={null} side={false} onFocus={() => {}} onReport={() => {}}
    onSessionChanged={() => {}} pulse={0} findPulse={0} onSettings={() => {}} needsProject={false}
    onOpenProject={() => {}} onKeepHere={() => {}} theme="dark" dockW={560} dockMax={880} onDockW={() => {}} />);
  const picker = await waitFor(() => {
    const button = container.querySelector('[data-action="plan.mode"]');
    if (!button || !listeners.size) throw new Error("Composer is not ready");
    return button;
  });
  fireEvent.click(picker);
  const choice = await waitFor(() => {
    const item = document.querySelector('[role="menu"] [data-value="agent_planner"]');
    if (!item) throw new Error("Agent + Planner is unavailable");
    return item;
  });
  fireEvent.click(choice);
  const box = container.querySelector<HTMLTextAreaElement>('textarea[aria-label="任务输入"]')!;
  const send = (text: string) => {
    fireEvent.change(box, { target: { value: text } });
    fireEvent.click(screen.getByRole("button", { name: "发送" }));
  };
  send("build it");
  await waitFor(() => expect(submit).toHaveBeenCalledWith("build it", undefined, "agent_planner", expect.any(String)));
  const selection = submit.mock.calls[0][3];
  await act(async () => listeners.forEach((listener) => listener({
    kind: "phase", code: "planner.approved", msgIndex: 1, plannerSelection: selection,
  })));
  expect(picker.textContent).toMatch(/Agent$/);
  send("check it");
  await waitFor(() => expect(submit).toHaveBeenLastCalledWith("check it", undefined));
});
