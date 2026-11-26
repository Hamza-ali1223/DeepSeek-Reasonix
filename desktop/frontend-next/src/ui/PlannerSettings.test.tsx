// @vitest-environment jsdom
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import "./testkit";
import { MockPort } from "../port/mock";
import { PlannerSettings } from "./PlannerSettings";

afterEach(cleanup);

it("saves the research rounds, the consultation switch, its limit, and the timeout together", async () => {
  const port = new MockPort();
  vi.spyOn(port, "mcp").mockResolvedValue({ servers: [], live: true, scope: { root: "/workspace", name: "workspace", key: "workspace", repo: false, overrides: 0 } });
  const save = vi.spyOn(port, "savePlannerSettings");
  render(<PlannerSettings port={port} />);
  const user = userEvent.setup();

  const rounds = await screen.findByRole("spinbutton", { name: "调研轮数" });
  await user.clear(rounds);
  await user.type(rounds, "8");

  await user.click(screen.getByRole("checkbox", { name: "允许咨询规划模型" }));
  const limit = screen.getByRole("spinbutton", { name: "每轮咨询上限" });
  await user.clear(limit);
  await user.type(limit, "2");
  const timeout = screen.getByRole("spinbutton", { name: "单次咨询超时" });
  await user.clear(timeout);
  await user.type(timeout, "240");

  await user.click(screen.getByRole("radio", { name: "自定义" }));
  await user.click(screen.getByRole("checkbox", { name: "grep" }));

  // One save writes every planner value: the footer is the only control that
  // reaches the kernel, so a user never has to guess which field committed.
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(save).toHaveBeenCalledWith({
    consultation: true,
    limit: 2,
    researchRounds: 8,
    consultationTimeoutSeconds: 240,
    tools: ["read_file"],
  }));
});

it("keeps a draft that failed to save", async () => {
  const port = new MockPort();
  vi.spyOn(port, "savePlannerSettings").mockRejectedValue(new Error("disk full"));
  render(<PlannerSettings port={port} />);
  const user = userEvent.setup();

  const rounds = await screen.findByRole("spinbutton", { name: "调研轮数" });
  await user.clear(rounds);
  await user.type(rounds, "7");
  await user.click(screen.getByRole("button", { name: "保存" }));

  // The value the user typed stays on screen: a reload would lose the one thing
  // the error is about.
  await waitFor(() => expect(screen.getByRole("alert")).toBeTruthy());
  expect((screen.getByRole("spinbutton", { name: "调研轮数" }) as HTMLInputElement).value).toBe("7");
});

it("reverts every planner value to the saved ones", async () => {
  const port = new MockPort();
  const save = vi.spyOn(port, "savePlannerSettings");
  render(<PlannerSettings port={port} />);
  const user = userEvent.setup();

  const rounds = await screen.findByRole("spinbutton", { name: "调研轮数" });
  await user.clear(rounds);
  await user.type(rounds, "5");
  await user.click(screen.getByRole("button", { name: "还原" }));

  expect((screen.getByRole("spinbutton", { name: "调研轮数" }) as HTMLInputElement).value).toBe("12");
  expect(save).not.toHaveBeenCalled();
});

it("refuses to save a value outside its range", async () => {
  const port = new MockPort();
  const save = vi.spyOn(port, "savePlannerSettings");
  render(<PlannerSettings port={port} />);
  const user = userEvent.setup();

  const rounds = await screen.findByRole("spinbutton", { name: "调研轮数" });
  await user.clear(rounds);
  await user.type(rounds, "13");

  expect(screen.getByRole("status").textContent).toContain("超出范围");
  expect((screen.getByRole("button", { name: "保存" }) as HTMLButtonElement).disabled).toBe(true);
  expect(save).not.toHaveBeenCalled();
});

// A cap below three has no room for the two reserved rounds, so the hint must
// stop promising them.
it("stops promising a reserve below a cap of three", async () => {
  const port = new MockPort();
  render(<PlannerSettings port={port} />);
  const user = userEvent.setup();

  const rounds = await screen.findByRole("spinbutton", { name: "调研轮数" });
  await user.clear(rounds);
  await user.type(rounds, "2");

  expect(screen.queryByText(/第 1 轮前收尾/)).toBeNull();
  expect(screen.getByText(/没有预留轮次/)).toBeTruthy();
});
