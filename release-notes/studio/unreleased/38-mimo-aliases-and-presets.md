## 修复

- 旧的 MiMo 模型名列表现已指向厂商当前计费的模型：`mimo-v2-pro` 解析为 `mimo-v2.6-pro`，`mimo-v2-omni` 解析为 `mimo-v2.6-flash`，因此仍在使用旧名称的配置能够取到价格。#37
- MiMo 预设不再提供已在 2026-10-21 下线的 `mimo-v2.5` 和 `mimo-v2.5-pro`，模型列表只保留 V2.6 系列。#37
- MiMo 预设现在声明 V2.6 系列支持图片输入，与厂商文档一致；此前唯一被标记为可读图的模型已经下线。#37
- 仍保持原有 V2.5 配置的用户会自动升级到 V2.6 目录，而自行调整过模型、窗口或默认值的配置不受影响。#37

---
<!-- lang:en -->

## Fixed

- The old MiMo model names now resolve to the models the vendor bills today: `mimo-v2-pro` to `mimo-v2.6-pro` and `mimo-v2-omni` to `mimo-v2.6-flash`, so a configuration still using the retired names gets a price again. #37
- MiMo presets no longer offer `mimo-v2.5` and `mimo-v2.5-pro`, which Xiaomi retired on 2026-10-21; the model lists now hold only the V2.6 series. #37
- MiMo presets declare the V2.6 series image-capable, matching the vendor's documentation; the only model previously marked for images had been retired. #37
- Configurations still holding the V2.5 catalog move up to V2.6 automatically, while anything with a curated model list, window or default is left alone. #37
