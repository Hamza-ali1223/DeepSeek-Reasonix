import DOMPurify from "dompurify";
import { generate, parse, walk, type WalkOptionsNoVisit } from "css-tree";

function safeCss(css: string, context: "stylesheet" | "declarationList", keyframes = false): string {
  const ast = parse(css, { context });
  const options: WalkOptionsNoVisit = { leave(node, item, list) {
    if ((node.type === "Atrule" && !(keyframes && node.name.toLowerCase() === "keyframes")) || node.type === "Raw") {
      if (item) list.remove(item);
    } else if (node.type === "Declaration") {
      let unsafe = false;
      walk(node.value, (value) => {
        if (value.type === "Raw" || (value.type === "Url" && !/^#[\w-]+$/.test(value.value))) unsafe = true;
      });
      if (unsafe && item) list.remove(item);
    }
  } };
  walk(ast, options);
  return generate(ast);
}

const SVG_NS = "http://www.w3.org/2000/svg";
const SMIL = ["animate", "animateMotion", "animateTransform", "set"];
const SMIL_ATTRS = ["attributeName", "attributeType", "begin", "dur", "end", "min", "max", "restart", "repeatCount", "repeatDur", "fill", "from", "to", "by", "values", "keyTimes", "keySplines", "keyPoints", "calcMode", "additive", "accumulate", "rotate", "path", "type"];
const LOCAL = /^#[\w.:-]+$/;
// An animation retargeted at one of these could navigate, fetch or script once it ran.
const UNSAFE_TARGET = /^(?:(?:xlink:)?href|xml:base|src|on\w*)$/i;

/** What a picture may keep beyond a still, self-contained drawing. Every allowance stays inert. */
export type SvgAllowance = { motion?: boolean; localReferences?: boolean };

/** The largest fence source a diagram renders; sanitizing and encoding run once per settled change. */
export const DIAGRAM_SOURCE_LIMIT = 50_000;

function svgRoot(source: string): Element {
  const parsed = new DOMParser().parseFromString(source, "image/svg+xml");
  const root = parsed.documentElement;
  if (parsed.querySelector("parsererror") || root.localName !== "svg") throw new Error("Invalid SVG output");
  if (root.namespaceURI === SVG_NS) return root;
  // An <svg> without xmlns is ordinary in HTML but is no SVG at all as an image document.
  if (root.namespaceURI !== null) throw new Error("Invalid SVG output");
  const scoped = new DOMParser().parseFromString(`<r xmlns="${SVG_NS}">${new XMLSerializer().serializeToString(root)}</r>`, "image/svg+xml");
  const svg = scoped.documentElement.firstElementChild;
  if (scoped.querySelector("parsererror") || svg?.namespaceURI !== SVG_NS || svg.localName !== "svg") throw new Error("Invalid SVG output");
  return svg;
}

/** A copy may only reach a drawing that holds no copies itself, so references cannot multiply or loop. */
function confineReferences(svg: Element) {
  const byId = new Map<string, Element>();
  for (const element of svg.querySelectorAll("[id]")) if (!byId.has(element.getAttribute("id")!)) byId.set(element.getAttribute("id")!, element);
  for (const element of [svg, ...svg.querySelectorAll("*")]) {
    for (const attr of [...element.attributes]) {
      if (attr.localName === "href" && (element.localName === "a" || !LOCAL.test(attr.value.trim()))) element.removeAttributeNode(attr);
    }
  }
  for (const use of [...svg.getElementsByTagNameNS(SVG_NS, "use")]) {
    const href = [...use.attributes].find((attr) => attr.localName === "href")?.value.trim();
    const target = href ? byId.get(href.slice(1)) : undefined;
    if (!target || target.localName === "use" || target.getElementsByTagNameNS(SVG_NS, "use").length > 0) use.remove();
  }
}

/** Chat diagrams are inert SVG images: no script, HTML, navigation or remote resources.
 *  By default also still and self-contained; `motion` keeps SMIL and CSS keyframes, and
 *  `localReferences` keeps `#id` references to shapes inside the same drawing. */
export function sanitizeInertSvg(source: string, allow: SvgAllowance = {}): string {
  const clean = DOMPurify.sanitize(svgRoot(source), {
    USE_PROFILES: { svg: true, svgFilters: true },
    ADD_TAGS: [...(allow.motion ? SMIL : []), ...(allow.localReferences ? ["use"] : [])],
    ADD_ATTR: allow.motion ? SMIL_ATTRS : [],
    FORBID_TAGS: ["foreignObject", "image", "feImage", ...(allow.motion ? [] : SMIL), ...(allow.localReferences ? [] : ["use", "mpath"])],
    FORBID_ATTR: ["src", "xml:base", ...(allow.localReferences ? [] : ["href", "xlink:href"])],
  });
  // DOMPurify serializes as HTML, which drops an xmlns that was never written as an attribute.
  const svg = svgRoot(clean);
  for (const animation of [...svg.querySelectorAll("*")].filter((element) => SMIL.includes(element.localName))) {
    // DOMPurify strips an unsafe attributeName but keeps the element and its to/values, so a missing target is refused too.
    const target = animation.getAttribute("attributeName")?.trim() ?? "";
    if (animation.localName !== "animateMotion" && (!target || UNSAFE_TARGET.test(target))) animation.remove();
  }
  if (allow.localReferences) confineReferences(svg);
  for (const style of svg.querySelectorAll("style")) {
    try { style.textContent = safeCss(style.textContent ?? "", "stylesheet", allow.motion); }
    catch { style.remove(); }
  }
  for (const element of [svg, ...svg.querySelectorAll("*")]) {
    for (const attr of [...element.attributes]) {
      if (attr.name === "style") {
        try { element.setAttribute("style", safeCss(attr.value, "declarationList")); }
        catch { element.removeAttribute("style"); }
      } else if (/url\s*\(/i.test(attr.value) && !/^url\(["']?#[\w-]+["']?\)$/.test(attr.value)) {
        element.removeAttribute(attr.name);
      }
    }
  }
  const bounds = svg.getAttribute("viewBox")?.trim().split(/[\s,]+/).map(Number);
  if (bounds?.length === 4 && bounds.every(Number.isFinite) && bounds[2]! > 0 && bounds[3]! > 0) {
    svg.setAttribute("width", String(bounds[2]));
    svg.setAttribute("height", String(bounds[3]));
  }
  return new XMLSerializer().serializeToString(svg);
}
