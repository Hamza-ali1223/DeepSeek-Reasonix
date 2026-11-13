## 新增

- Studio 将 ```` ```svg ```` 围栏显示为图片，支持预览／源代码切换和复制；流式尾块完成前保留源代码。fork #1
- SVG 只以净化后的静态 `<img>` 显示，不执行脚本、不打开链接、不加载远程资源；无效或超过 50,000 字符的 SVG 显示原因并保留源代码。fork #1
- SVG 支持 SMIL 与 CSS `@keyframes` 动画，系统开启“减少动态效果”时显示静止画面；`<use>` 可复用同一图中的图形。Mermaid 图表保持静止。fork #1
- SVG 与 Mermaid 图片支持 Ctrl + 滚轮（或触控板双指）以指针为中心缩放、拖动平移、放大／缩小／复原按钮、双击复原，并可在共用的图片查看器中全屏查看。fork #1

---
<!-- lang:en -->

## New

- Studio draws ```` ```svg ```` fences as pictures with Preview/Source and Copy source. Streaming tails remain source until settled. fork #1
- SVG is shown only as a sanitized inert `<img>`: no scripts, navigation or remote loads. Invalid SVG, or a fence over 50,000 characters, shows the reason beside its source. fork #1
- SVG animates with SMIL and CSS `@keyframes`, and shows a still drawing when the system asks for reduced motion. `<use>` reuses shapes from the same drawing. Mermaid diagrams stay still. fork #1
- SVG and Mermaid pictures zoom with Ctrl + scroll (or a trackpad pinch) around the pointer, pan by dragging, step with zoom in/out/reset, reset on double-click, and open full screen in the shared image viewer. fork #1
