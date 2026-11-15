本版在工作区之外写入或递归删除时改为向你确认，受保护路径仍然直接拒绝。

## 新增

- 文件工具要写到工作区之外时会出现卡片：只允许这一次、本会话允许该文件夹，或把该文件夹记录到项目中。fork #23
- 在工作区和已授权文件夹之外递归删除时会出现卡片，答复只对这一条命令有效。fork #23
- 卡片不会被 Ask、Auto、Yolo 模式、记住的选择或钩子代答；无人可问时（无人值守运行、无法询问的子代理）保持原来的拒绝。fork #23
- 终端中的卡片写明涉及的文件夹，删除被拒绝时会说明原因。fork #23

## 升级须知

- 工作区之外的写入与递归删除不再一律拒绝，改由你在卡片上决定；家目录、文件系统根、工作区祖先、Reasonix 自身存储与凭据目录仍然直接拒绝。fork #23
- 从写入卡片选择“记录到项目”的文件夹会成为项目的可写文件夹，因此在该文件夹内递归删除不再出现卡片（见 #22）。fork #23

---
<!-- lang:en -->

## New

- A file tool writing outside the workspace raises a card: this write only, this folder for the session, or this folder recorded for the project. fork #23
- A recursive delete outside the workspace and its granted folders raises a card, and the answer covers that one command. fork #23
- No posture (Ask, Auto, Yolo), remembered choice or hook answers the card; with nobody to ask (a headless run, a sub-agent that cannot ask) the original refusal stands. fork #23
- The terminal names the folder on the card, and a refused delete says why. fork #23

## Upgrade notes

- Writes and recursive deletes outside the workspace are no longer refused outright; you decide on the card. Home directories, filesystem roots, workspace ancestors, Reasonix's own storage and credential directories stay refused. fork #23
- A folder recorded for the project from a write card becomes one of its writable folders, so a recursive delete inside it raises no card (#22). fork #23
