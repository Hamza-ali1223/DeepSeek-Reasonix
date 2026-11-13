// @vitest-environment jsdom
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { cleanup, fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./testkit";
import { Markdown } from "./Markdown";

const engine = vi.hoisted(() => ({ render: vi.fn(async () => '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100"><text>Browser</text></svg>') }));
vi.mock("./mermaid", () => ({ renderMermaid: engine.render }));

Element.prototype.getAnimations ??= () => [];
beforeEach(() => { HTMLElement.prototype.setPointerCapture = vi.fn(); });
afterEach(() => { cleanup(); vi.clearAllMocks(); });

const svgFence = "```svg\n" + '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100"><circle cx="50" cy="50" r="40" fill="teal"/></svg>\n' + "```";
const viewOf = (c: HTMLElement) => c.querySelector<HTMLElement>(".diagram-preview")!;
const imgOf = (c: HTMLElement) => c.querySelector<HTMLImageElement>(".diagram-preview img")!;
const scaleOf = (c: HTMLElement) => Number(/scale\(([\d.]+)\)/.exec(imgOf(c).style.transform)?.[1] ?? 1);
const panOf = (c: HTMLElement) => /translate\((-?[\d.]+)px, (-?[\d.]+)px\)/.exec(imgOf(c).style.transform)?.slice(1).map(Number) ?? [0, 0];

it("zooms with Ctrl + wheel over the picture, and leaves a plain wheel to scroll the chat", () => {
  const { container } = render(<Markdown text={svgFence} />);
  expect(fireEvent.wheel(viewOf(container), { deltaY: -100 })).toBe(true);
  expect(scaleOf(container)).toBe(1);
  expect(fireEvent.wheel(viewOf(container), { deltaY: -100, ctrlKey: true })).toBe(false);
  const zoomed = scaleOf(container);
  expect(zoomed).toBeGreaterThan(1);
  fireEvent.wheel(viewOf(container), { deltaY: 100, ctrlKey: true });
  expect(scaleOf(container)).toBeLessThan(zoomed);
  for (let i = 0; i < 60; i++) fireEvent.wheel(viewOf(container), { deltaY: -100, ctrlKey: true });
  expect(scaleOf(container)).toBe(8);
  for (let i = 0; i < 60; i++) fireEvent.wheel(viewOf(container), { deltaY: 100, ctrlKey: true });
  expect(scaleOf(container)).toBe(0.2);
});

it("steps with the buttons, shows the level, and resets by button or double-click", async () => {
  const user = userEvent.setup();
  const { container } = render(<Markdown text={svgFence} />);
  await user.click(screen.getByRole("button", { name: "放大" }));
  expect(scaleOf(container)).toBe(1.25);
  const reset = screen.getByRole("button", { name: "恢复原始大小" });
  expect(reset.textContent).toBe("125%");
  await user.click(screen.getByRole("button", { name: "缩小" }));
  await user.click(screen.getByRole("button", { name: "缩小" }));
  expect(scaleOf(container)).toBe(0.8);
  await user.click(reset);
  expect(scaleOf(container)).toBe(1);
  expect(reset.textContent).toBe("100%");
  await user.click(screen.getByRole("button", { name: "放大" }));
  await user.dblClick(imgOf(container));
  expect(scaleOf(container)).toBe(1);
});

it("drags the picture to move it, and stops when the button is released", () => {
  const { container } = render(<Markdown text={svgFence} />);
  fireEvent.pointerDown(imgOf(container), { pointerId: 1, button: 0, pointerType: "mouse", clientX: 10, clientY: 10 });
  fireEvent.pointerMove(viewOf(container), { pointerId: 1, clientX: 40, clientY: 25 });
  expect(panOf(container)).toEqual([30, 15]);
  fireEvent.pointerUp(viewOf(container), { pointerId: 1 });
  fireEvent.pointerMove(viewOf(container), { pointerId: 1, clientX: 90, clientY: 90 });
  expect(panOf(container)).toEqual([30, 15]);
});

it("opens the picture full screen in the shared image viewer, and returns focus on Escape", async () => {
  const user = userEvent.setup();
  const { container } = render(<Markdown text={svgFence} />);
  const open = screen.getByRole("button", { name: "全屏查看" });
  await user.click(open);
  const dialog = screen.getByRole("dialog", { name: "SVG 图片" });
  expect(container.contains(dialog)).toBe(false);
  expect(within(dialog).getByRole("img").getAttribute("src")).toBe(imgOf(container).getAttribute("src"));
  await user.keyboard("{Escape}");
  expect(screen.queryByRole("dialog")).toBeNull();
  expect(document.activeElement).toBe(open);
});

it("keeps the full-screen viewer open when a drag ends on its backdrop, and still closes on a plain click there", async () => {
  const user = userEvent.setup();
  render(<Markdown text={svgFence} />);
  await user.click(screen.getByRole("button", { name: "全屏查看" }));
  const dialog = screen.getByRole("dialog", { name: "SVG 图片" });
  const picture = within(dialog).getByRole("img");
  const view = picture.parentElement!;
  await user.click(within(dialog).getByRole("button", { name: "放大" }));
  fireEvent.pointerDown(picture, { pointerId: 1, button: 0, clientX: 100, clientY: 100 });
  fireEvent.pointerMove(view, { pointerId: 1, clientX: 100, clientY: 260 });
  fireEvent.pointerUp(view, { pointerId: 1, clientX: 100, clientY: 260 });
  fireEvent.click(view, { detail: 1, clientX: 100, clientY: 260 });
  expect(screen.getByRole("dialog")).toBe(dialog);
  expect(picture.style.transform).toBe("translate(0px, 160px) scale(1.25)");
  fireEvent.pointerDown(view, { pointerId: 2, button: 0, clientX: 5, clientY: 5 });
  fireEvent.pointerUp(view, { pointerId: 2, clientX: 5, clientY: 5 });
  fireEvent.click(view, { detail: 1, clientX: 5, clientY: 5 });
  expect(screen.queryByRole("dialog")).toBeNull();
});

it("gives Mermaid diagrams the same controls", async () => {
  const { container } = render(<Markdown text={"```mermaid\nflowchart TD\n  A --> B\n```"} />);
  await screen.findByRole("img", { name: "Mermaid 图表" });
  expect(fireEvent.wheel(viewOf(container), { deltaY: -100, ctrlKey: true })).toBe(false);
  expect(scaleOf(container)).toBeGreaterThan(1);
  expect(screen.getByRole("button", { name: "全屏查看" })).toBeTruthy();
});

it("offers no zoom while the fence is still arriving or showing source, and starts again at 100%", async () => {
  const user = userEvent.setup();
  const { container, rerender } = render(<Markdown text={svgFence} streaming />);
  expect(screen.queryByRole("button", { name: "放大" })).toBeNull();
  rerender(<Markdown text={svgFence} />);
  await user.click(screen.getByRole("button", { name: "放大" }));
  await user.click(screen.getByRole("button", { name: "源代码" }));
  expect(screen.queryByRole("button", { name: "放大" })).toBeNull();
  expect(screen.queryByRole("button", { name: "全屏查看" })).toBeNull();
  await user.click(screen.getByRole("button", { name: "预览" }));
  expect(scaleOf(container)).toBe(1);
});
