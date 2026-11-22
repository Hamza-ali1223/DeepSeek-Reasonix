// @vitest-environment jsdom
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import "./testkit";
import { MockPort } from "../port/mock";
import { PlannerSettings } from "./PlannerSettings";

afterEach(cleanup);

it("saves consultation, its per-turn limit, and planner-only research tools", async () => {
  const port = new MockPort();
  const save = vi.spyOn(port, "savePlannerSettings");
  render(<PlannerSettings port={port} />);
  const user = userEvent.setup();
  await user.click(await screen.findByRole("checkbox", { name: "允许咨询规划模型" }));
  const limit = screen.getByRole("spinbutton", { name: "每轮咨询上限" });
  await user.clear(limit);
  await user.type(limit, "2");
  await user.type(screen.getByRole("textbox", { name: "规划模型研究工具" }), "read_file{Enter}mcp-tool:docs/search");
  await user.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(save).toHaveBeenCalledWith({
    consultation: true, limit: 2, tools: ["read_file", "mcp-tool:docs/search"],
  }));
});
