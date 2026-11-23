## 新增

- 在 Studio 输入框选择「Agent + Planner」，先由专用规划模型拟定方案，再由执行模型按审批设置执行。fork #34
- 可为新会话开启规划咨询，设置每轮上限，并选择规划模型可使用的研究工具。fork #34
- 会话概览分别显示规划与执行模型的用量，以及服务商报告的缓存用量。fork #34

## 修复

- 方案审批记录跟随原始请求。即使重启时只保存了请求，恢复后也只保留一份审批记录。fork #34

---
<!-- lang:en -->

## New

- Choose Agent + Planner in the Studio composer to get a dedicated plan before the executor starts under your approval settings. fork #34
- Enable planner consultations for new sessions, set the per-turn limit, and choose the planner's research tools. fork #34
- See planner and executor usage separately in Session Overview, including cached tokens when reported. fork #34

## Fixed

- Keep one plan approval receipt with its request after a restart, even if the assistant reply was never saved. fork #34
