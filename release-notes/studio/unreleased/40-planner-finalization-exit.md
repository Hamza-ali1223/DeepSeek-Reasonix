## 修复

- 规划轮次的调研轮数用满后，规划模型仍可以提交计划或得出无需修改的结论，不再被整轮拒绝后丢弃。
- 在更早轮次已提交的有效计划，不会再因为同一轮后续调研耗尽预算而被回滚丢弃。#40

---

<!-- lang:en -->

## Fixed

- A planning turn that spent its research rounds can now submit its plan or conclude that nothing needs changing, instead of having every call in that round refused and the turn discarded.
- A valid plan submitted in an earlier round is no longer discarded by a later round that exhausts the research budget. #40
