import { useEffect, useState } from "react";
import { t } from "../i18n";
import { reason } from "../i18n/kernel";
import type { PlannerSettings as Settings } from "../port/boundary";
import type { AgentPort } from "../port/port";
import { SettingRow } from "./SettingRow";

export function PlannerSettings({ port }: { port: AgentPort }) {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [saved, setSaved] = useState<Settings | null>(null);
  const [toolsText, setToolsText] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    let current = true;
    void port.plannerSettings().then((value) => {
      if (!current) return;
      setSettings(value);
      setSaved(value);
      setToolsText(value.tools.join("\n"));
    }).catch((err: unknown) => {
      if (current) setError(reason(err));
    });
    return () => { current = false; };
  }, [port]);
  if (!settings) return error ? <div role="alert">{error}</div> : null;

  const tools = toolsText.split(/\r?\n/).map((tool) => tool.trim()).filter(Boolean);
  const save = async () => {
    setBusy(true);
    setError("");
    try {
      const value = await port.savePlannerSettings({ ...settings, tools });
      setSaved(value);
      setSettings(value);
      setToolsText(value.tools.join("\n"));
    } catch (err) {
      setError(reason(err));
    } finally {
      setBusy(false);
    }
  };
  const validLimit = Number.isInteger(settings.limit) && settings.limit >= 1 && settings.limit <= 8;
  const changed = settings.consultation !== saved?.consultation || settings.limit !== saved?.limit
    || tools.join("\n") !== saved?.tools.join("\n");
  return (
    <>
      <SettingRow id="planner-consultation" title={t("settings.planner_consultation")} desc={t("settings.planner_consultation_hint")}>
        <input aria-label={t("settings.planner_consultation")} type="checkbox" checked={settings.consultation} disabled={busy}
          onChange={(e) => setSettings({ ...settings, consultation: e.target.checked })} />
      </SettingRow>
      <SettingRow id="planner-consultation-limit" title={t("settings.planner_consultation_limit")}>
        <input aria-label={t("settings.planner_consultation_limit")} type="number" min={1} max={8} value={settings.limit}
          disabled={busy || !settings.consultation} onChange={(e) => setSettings({ ...settings, limit: Number(e.target.value) })} />
      </SettingRow>
      <SettingRow id="planner-tools" title={t("settings.planner_tools")} desc={t("settings.planner_tools_hint")}>
        <textarea aria-label={t("settings.planner_tools")} rows={3} value={toolsText} disabled={busy}
          onChange={(e) => setToolsText(e.target.value)} />
        <button type="button" className="spill" disabled={busy || !changed || !validLimit} onClick={() => void save()}>
          {t("common.save")}
        </button>
      </SettingRow>
      {error && <div role="alert">{error}</div>}
    </>
  );
}
