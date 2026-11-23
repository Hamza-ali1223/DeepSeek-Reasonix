## 修复

- 价格表的 90 天新鲜度检查不再拿已下线的模型当过期行处理；MiMo-V2.5-Pro 和 MiMo-V2.5 已于 2026-10-21 下线，MiMo-V2-Flash 更早，已于 2026-06-30 下线。#31
- 这几行现在按下线日期结束有效期，不再作为当前价格提供，也不再让 CI 变红。#31
- 旧用量的计费不受影响：下线行仍保留在价格目录里，历史记录仍能按当时的价格结算。#31
- LongCat 的两个币种价格已重新核对厂商页面，核对日期一并更新。#31

---
<!-- lang:en -->

## Fixed

- The price table's 90-day freshness check no longer treats retired models as stale rows; MiMo-V2.5-Pro and MiMo-V2.5 were retired on 2026-10-21, and MiMo-V2-Flash earlier, on 2026-06-30. #31
- Those rows now end their effective lifetime on the retirement date instead of being offered as a current price, and no longer turn CI red. #31
- Usage that was already billed is unaffected: the retired rows stay in the catalogue, so history still resolves at the price it was charged at. #31
- LongCat's two currencies were re-read from the vendor's page and their check date moved forward. #31
