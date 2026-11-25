## 新功能

- 规划轮次的调研轮数现在可以配置，默认 12 轮，允许 1 到 12。规划模型会在每轮开头看到本轮上限和建议收尾的轮次，并在只剩两轮时收到一次提醒。上限仍由宿主强制，建议轮次只是提示。
- 规划模型现在知道一轮调研是什么：一次回复里的一批工具调用算一轮，不论那批里有几个调用。
- 咨询规划模型的单次超时现在可以配置，默认 90 秒，允许 15 到 600 秒。
- 每轮允许的咨询次数上限从 8 提高到 20。
- 规划轮次在预算用尽且没有产出计划时，会单独报告这一结果，不再和"还有一轮收尾"混为一谈。

## 清理

- 移除了不可达的 `PlannerDepthLight` 及其 2 轮预算，以及规划提示词里对应的半段说明。快速路由现在自带默认轮数，不再可能得出无限轮的决定。#41

---

<!-- lang:en -->

## Added

- The research rounds one planning turn may use are now configurable: default 12, range 1 to 12. The planner is told the cap and the round to aim to finish by at the head of the turn, and gets one reminder when two rounds remain. The host still enforces the cap; the target is guidance only.
- The planner now knows what a round is: one reply that issues a batch of tool calls, however many calls that batch holds.
- One consultation's timeout is now configurable: default 90 seconds, range 15 to 600 seconds.
- The per-turn consultation limit rises from 8 to 20.
- A planning turn that spends its budget without producing a plan reports that outcome on its own, instead of sharing one identity with "one finalization reply is still available".

## Clean-up

- Removed the unreachable `PlannerDepthLight`, its 2-round budget, and the matching half of the planner prompt. Fast routing now carries the default round count, so a routing decision can no longer mean "unlimited rounds". #41
