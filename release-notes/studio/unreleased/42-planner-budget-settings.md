## 新增

- 规划轮次的调研轮数现在可以配置，默认 12 轮，允许 1 到 12，规划模型会被告知上限和建议收尾的轮次。#41
- 咨询规划模型的单次超时现在可以配置，默认 90 秒，允许 15 到 600 秒，每轮咨询上限从 8 提高到 20。#41
- 规划轮次在预算用尽且没有产出计划时，会单独报告这一结果，不再和"还有一轮收尾"共用一个标识。#41

## 移除

- 移除了不可达的 `PlannerDepthLight` 及其 2 轮预算，以及规划提示词里对应的半段说明。#41

---

<!-- lang:en -->

## New

- The research rounds one planning turn may use are now configurable: default 12, range 1 to 12. The planner is told the cap and the round to aim to finish by. #41
- One consultation's timeout is now configurable, default 90 seconds, range 15 to 600, and the per-turn consultation limit rises from 8 to 20. #41
- A planning turn that spends its budget without producing a plan reports that outcome on its own, instead of sharing one identity with "one finalization reply is still available". #41

## Removed

- Removed the unreachable `PlannerDepthLight`, its 2-round budget, and the matching half of the planner prompt. #41
