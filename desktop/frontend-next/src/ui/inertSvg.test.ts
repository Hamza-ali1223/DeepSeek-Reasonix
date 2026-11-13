// @vitest-environment jsdom
import { expect, it } from "vitest";
import { sanitizeInertSvg } from "./inertSvg";

it("preserves real diagram geometry, labels, themes and local arrow markers", () => {
  const svg = sanitizeInertSvg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 100"><style>#graph .node{fill:#eee;stroke:#333} .edge{marker-end:url(#arrow)}</style><defs><marker id="arrow"><path d="M0 0 L5 5"/></marker></defs><g class="node" transform="translate(20,30)"><rect width="100" height="40"/><text>Browser</text></g><path class="edge" d="M0 0 L100 100"/></svg>');
  expect(svg).toContain('viewBox="0 0 200 100"');
  expect(svg).toContain('width="200"');
  expect(svg).toContain('height="100"');
  expect(svg).toContain('marker-end:url(#arrow)');
  expect(svg).toContain('translate(20,30)');
  expect(svg).toContain("Browser");
});

it("removes scripts, handlers, HTML, links, external images and CSS resource loads", () => {
  const svg = sanitizeInertSvg(`<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)">
    <script>alert(1)</script><foreignObject><iframe src="https://evil.test"/></foreignObject>
    <image href="https://evil.test/leak"/><use href="https://evil.test/icons.svg#x"/>
    <a href="javascript:alert(1)" onclick="alert(1)"><text>label</text></a>
    <a href="https://evil.test/"><text>web</text></a>
    <style>@import "https://evil.test/a.css";.x{fill:url(https://evil.test/a);stroke:red}.y{fill:u\\72l(https://evil.test/b)}</style>
    <rect style="fill:url(https://evil.test/c);stroke:blue" onmouseover="alert(1)"/>
    <animate attributeName="href" to="https://evil.test/"/>
  </svg>`);
  expect(svg).not.toMatch(/evil|javascript|onload|onclick|onmouseover|<script|foreignObject|<image|<use|<animate|@import/i);
  expect(svg).toContain("label");
  expect(svg).toContain("stroke:blue");
});

it("keeps the SVG namespace, which an image needs, whether or not the source wrote it", () => {
  for (const source of ['<svg viewBox="0 0 4 4"><rect width="1" height="1"/></svg>', '<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>']) {
    const svg = new DOMParser().parseFromString(sanitizeInertSvg(source), "image/svg+xml").documentElement;
    expect(svg.namespaceURI).toBe("http://www.w3.org/2000/svg");
    expect(svg.querySelector("rect")?.namespaceURI).toBe("http://www.w3.org/2000/svg");
  }
});

const NS = 'xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"';
const PICTURE = { motion: true, localReferences: true };

it("keeps the default profile still and self-contained, which is what Mermaid diagrams use", () => {
  const svg = sanitizeInertSvg(`<svg ${NS}><style>@keyframes k{to{opacity:0}}</style><path id="p" d="M0 0"/><use href="#p"/><circle r="1"><animate attributeName="r" values="1;2" dur="1s"/></circle></svg>`);
  expect(svg).not.toMatch(/<use|<animate|<set|@keyframes/);
});

it("keeps SMIL and CSS keyframe animation when motion is allowed", () => {
  const svg = sanitizeInertSvg(`<svg ${NS} viewBox="0 0 100 100"><style>@keyframes spin{from{transform:rotate(0)}to{transform:rotate(360deg)}} .s{animation:spin 2s infinite}</style>
    <circle class="s" r="5"><animate attributeName="r" values="5;10;5" dur="2s" repeatCount="indefinite"/></circle>
    <g><animateTransform attributeName="transform" type="rotate" from="0 50 50" to="360 50 50" dur="8s" repeatCount="indefinite"/></g>
    <path id="orbit" d="M0 0 L10 10"/><circle r="2"><animateMotion dur="3s" repeatCount="indefinite"><mpath href="#orbit"/></animateMotion></circle>
    <rect width="4" height="4"><set attributeName="fill" to="red" begin="1s"/></rect></svg>`, PICTURE);
  expect(svg).toContain("@keyframes spin");
  expect(svg).toContain("animation:spin 2s infinite");
  const doc = new DOMParser().parseFromString(svg, "image/svg+xml");
  const attrs = (tag: string, ...names: string[]) => {
    const el = [...doc.querySelectorAll("*")].find((e) => e.localName === tag);
    return names.map((n) => el?.getAttribute(n));
  };
  expect(attrs("animate", "attributeName", "values", "dur", "repeatCount")).toEqual(["r", "5;10;5", "2s", "indefinite"]);
  expect(attrs("animateTransform", "type", "from", "to", "dur")).toEqual(["rotate", "0 50 50", "360 50 50", "8s"]);
  expect(attrs("animateMotion", "dur")).toEqual(["3s"]);
  expect(attrs("mpath", "href")).toEqual(["#orbit"]);
  expect(attrs("set", "attributeName", "to", "begin")).toEqual(["fill", "red", "1s"]);
});

it("draws every local copy of a shape, whichever way the reference is spelled", () => {
  const svg = sanitizeInertSvg(`<svg ${NS}><defs><ellipse id="petal" rx="4" ry="10"/></defs><use href="#petal"/><use xlink:href="#petal" transform="rotate(30)"/></svg>`, { localReferences: true });
  const doc = new DOMParser().parseFromString(svg, "image/svg+xml");
  expect(doc.querySelectorAll("use")).toHaveLength(2);
  expect([...doc.querySelectorAll("use")].map((u) => u.getAttribute("href") ?? u.getAttributeNS("http://www.w3.org/1999/xlink", "href"))).toEqual(["#petal", "#petal"]);
});

it("drops what an animation or a reference could use to navigate, script or fetch", () => {
  const svg = sanitizeInertSvg(`<svg ${NS}>
    <a><animate attributeName="href" to="javascript:alert(1)" begin="0s"/><set attributeName="xlink:href" to="https://evil.test/"/><text>label</text></a>
    <rect width="4" height="4"><set attributeName="onclick" to="alert(1)"/><animate attributeName="fill" values="red;url(https://evil.test/p)" dur="1s"/></rect>
    <use href="https://evil.test/x.svg#a"/><use xlink:href="other.svg#b"/>
    <style>@import "https://evil.test/a.css"; @font-face{font-family:x;src:url(https://evil.test/f)} @keyframes k{to{fill:url(https://evil.test/k)}}</style>
  </svg>`, PICTURE);
  expect(svg).not.toMatch(/evil|javascript|onclick|other\.svg|@import|@font-face|attributeName="(xlink:)?href"/i);
  expect(svg).toContain("label");
  expect(svg).toContain("@keyframes k");
});

it("lets a reference reach a shape but not another reference, so copies cannot multiply", () => {
  const svg = sanitizeInertSvg(`<svg ${NS}><defs><ellipse id="petal" rx="4" ry="10"/><g id="pair"><use href="#petal"/><use href="#petal" transform="rotate(30)"/></g></defs><use href="#petal"/><use href="#pair"/><g id="loop"><use href="#loop"/></g></svg>`, PICTURE);
  const hrefs = [...new DOMParser().parseFromString(svg, "image/svg+xml").querySelectorAll("use")].map((u) => u.getAttribute("href"));
  expect(hrefs).toEqual(["#petal", "#petal", "#petal"]);
});

it("fails closed for non-SVG or malformed output", () => {
  expect(() => sanitizeInertSvg("<html><body>not svg</body></html>")).toThrow();
  expect(() => sanitizeInertSvg('<svg><g></svg>')).toThrow();
});
