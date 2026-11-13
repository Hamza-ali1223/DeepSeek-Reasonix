// @vitest-environment jsdom
import { afterEach, expect, it, vi } from "vitest";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./testkit";
import { Markdown } from "./Markdown";

afterEach(() => { cleanup(); vi.restoreAllMocks(); });
const fence = (s: string) => "```svg\n" + s + "```";
const pictureOf = (container: HTMLElement) => {
  const src = container.querySelector<HTMLImageElement>(".diagram-preview img")?.getAttribute("src") ?? "";
  expect(src.startsWith("data:image/svg+xml;charset=utf-8,")).toBe(true);
  return decodeURIComponent(src.slice(src.indexOf(",") + 1));
};
const circle = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100"><circle cx="50" cy="50" r="40" fill="teal"/><text x="100" y="55">Hello</text></svg>\n';

it("routes only SVG fences to an image preview and keeps source and copy available", async () => {
  const user = userEvent.setup();
  const write = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
  const { container } = render(<Markdown text={fence(circle) + "\n\n```go\nfunc main() {}\n```\n\n**Next paragraph**"} />);
  await waitFor(() => expect(container.querySelector('.code-wrap[data-lang="go"] [class*="hljs-"]')).toBeTruthy());
  expect(screen.getByRole("img", { name: "SVG 图片" })).toBeTruthy();
  expect(pictureOf(container)).toContain('fill="teal"');
  expect(container.querySelector(".diagram-preview svg")).toBeNull();
  await user.click(screen.getByRole("button", { name: "源代码" }));
  expect(container.querySelector(".diagram-source")?.textContent).toBe(circle);
  await user.click(screen.getByRole("button", { name: "复制 SVG 源代码" }));
  expect(write).toHaveBeenCalledWith(circle);
  await user.click(screen.getByRole("button", { name: "预览" }));
  expect(screen.getByRole("img", { name: "SVG 图片" })).toBeTruthy();
  expect(container.querySelector('.code-wrap[data-lang="go"] .code-tx')?.textContent).toBe("func main() {}");
  expect(container.querySelector("strong")?.textContent).toBe("Next paragraph");
});

it("draws only an inert picture: scripts, handlers, HTML and remote loads never survive", () => {
  const hostile = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" onload="alert(1)"><script>alert(1)</script><foreignObject><iframe src="https://evil.test"/></foreignObject><image href="https://evil.test/leak"/><a href="javascript:alert(1)"><text>label</text></a><rect style="fill:url(https://evil.test/c);stroke:blue"/></svg>\n`;
  const { container } = render(<Markdown text={fence(hostile)} />);
  const svg = pictureOf(container);
  expect(svg).not.toMatch(/evil|javascript|onload|<script|foreignObject|<image/i);
  expect(svg).toContain("label");
  expect(container.querySelector("script, iframe, foreignObject")).toBeNull();
});

it("draws an SVG written without its namespace, as a browser page would", () => {
  const { container } = render(<Markdown text={fence('<svg viewBox="0 0 20 10"><rect width="5" height="5"/></svg>\n')} />);
  expect(pictureOf(container)).toContain('xmlns="http://www.w3.org/2000/svg"');
});

it("keeps invalid SVG readable as source with the reason, and recovers when the source changes", () => {
  const { container, rerender } = render(<Markdown text={fence("<svg><g></svg>\n")} />);
  expect(screen.getByRole("alert").textContent).toContain("SVG 预览失败");
  expect(container.querySelector(".diagram-source")?.textContent).toBe("<svg><g></svg>\n");
  expect(container.querySelector(".diagram-preview")).toBeNull();
  rerender(<Markdown text={fence(circle)} />);
  expect(screen.getByRole("img", { name: "SVG 图片" })).toBeTruthy();
});

it("refuses a fence too large to preview and keeps it readable", () => {
  const huge = `<svg xmlns="http://www.w3.org/2000/svg">${"<rect/>".repeat(9000)}</svg>\n`;
  const { container } = render(<Markdown text={fence(huge)} />);
  expect(screen.getByRole("alert").textContent).toContain("50,000");
  expect(container.querySelector(".diagram-preview")).toBeNull();
  expect(container.querySelector(".diagram-source")?.textContent).toBe(huge);
});

const spinning = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><ellipse id="petal" rx="4" ry="10"/></defs><use href="#petal"/><g><animateTransform attributeName="transform" type="rotate" from="0 50 50" to="360 50 50" dur="8s" repeatCount="indefinite"/></g></svg>\n';

function motionSetting(reduce: boolean) {
  const listeners = new Set<() => void>();
  const state = { reduce };
  vi.spyOn(window, "matchMedia").mockImplementation((query: string) => ({
    get matches() { return query.includes("prefers-reduced-motion: reduce") && state.reduce; },
    media: query, onchange: null, addListener() {}, removeListener() {}, dispatchEvent: () => false,
    addEventListener: (_: string, f: () => void) => { listeners.add(f); },
    removeEventListener: (_: string, f: () => void) => { listeners.delete(f); },
  }) as unknown as MediaQueryList);
  return (next: boolean) => act(() => { state.reduce = next; listeners.forEach((f) => f()); });
}

it("plays SMIL animation and keeps local references in the picture", () => {
  motionSetting(false);
  const { container } = render(<Markdown text={fence(spinning)} />);
  const svg = pictureOf(container);
  expect(svg).toContain("<animateTransform");
  expect(svg).toContain('href="#petal"');
});

it("shows a still picture while the system asks for reduced motion, and follows the setting live", () => {
  const flip = motionSetting(true);
  const { container } = render(<Markdown text={fence(spinning)} />);
  expect(pictureOf(container)).not.toContain("<animateTransform");
  expect(pictureOf(container)).toContain('href="#petal"');
  flip(false);
  expect(pictureOf(container)).toContain("<animateTransform");
  flip(true);
  expect(pictureOf(container)).not.toContain("<animateTransform");
});

it("shows the live tail as source, and draws it once the message settles", () => {
  const { container, rerender } = render(<Markdown text={fence('<svg xmlns="http://www.w3.org/2000/svg"><circle r="4"')} streaming />);
  expect(container.querySelector(".diagram-preview")).toBeNull();
  expect(container.querySelector("[role=alert]")).toBeNull();
  expect(container.querySelector(".diagram-source")?.textContent).toContain("<circle");
  expect(screen.queryByRole("button", { name: "复制 SVG 源代码" })).toBeNull();
  rerender(<Markdown text={fence(circle)} />);
  expect(screen.getByRole("img", { name: "SVG 图片" })).toBeTruthy();
});
