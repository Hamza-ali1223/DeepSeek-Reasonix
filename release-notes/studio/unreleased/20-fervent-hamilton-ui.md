## 新增

- 设置侧栏按类别分组，页面索引按卡片列出；分隔线与开关统一为一种样式。fork #20
- 工作台“关闭全部标签”会关闭所有标签（包括文件）；文件标签也有同样的菜单；有未保存修改的文件会保留，并提示保留了几个。fork #20
- 侧栏中运行中的会话只显示一个旋转标记，不再同时显示圆点和“运行中”。fork #20

## 修复

- 单独成行的公式按独立公式显示，`\tag` 编号可以正常渲染。fork #20
- 关闭工作台后再次打开，会回到启动页而不是文件列表。fork #20
- 中转服务在返回结束标记后仍保持连接时，回复会在短暂等待后正常结束。fork #20
- Patina 主题在浅色和深色下，选中项都能与背景区分。fork #20
- 扩展范围的项目列表在设置栏内展开，不再超出左边缘。fork #20
- Windows 上设置标题栏的最小化、最大化和关闭按钮可以点击。fork #20

## 移除

- 移除左侧图标栏顶部的会话图标；“返回工作区”或 Esc 可离开设置，顶部切换按钮可显示或隐藏会话列表。fork #20

---
<!-- lang:en -->

## New

- The settings sidebar is grouped by category, and the page index lists cards; dividers and switches share one style. fork #20
- "Close all tabs" in the workbench closes every tab, files included; file tabs offer the same menu; a file with unsaved edits is kept, and the strip says how many were kept. fork #20
- A running session in the sidebar shows one spinning mark instead of a dot and the word. fork #20

## Fixed

- A formula alone on its line renders as display math, so `\tag` numbers render. fork #20
- Reopening the workbench after closing it lands on the launcher, not the file list. fork #20
- A reply ends shortly after its finish marker when a relay keeps the connection open. fork #20
- Patina's selected items stand apart from their background in light and dark. fork #20
- The extension-scope project list opens inside the settings column instead of past its left edge. fork #20
- The minimise, maximise and close buttons on the settings title bar can be clicked on Windows. fork #20

## Removed

- The conversation icon at the top of the left icon strip is gone; "Back to workspace" or Escape leaves Settings, and the top toggle shows or hides the session list. fork #20
