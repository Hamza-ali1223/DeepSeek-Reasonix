import { useEffect, useState } from "react";
import { t } from "../i18n";
import { reason } from "../i18n/kernel";
import type { PlannerSettings as Settings } from "../port/boundary";
import type { AgentPort } from "../port/port";
import { PillButton, SettingRow } from "./SettingRow";

// The research cap and the timeout are two numbers beside two switches. One
// save writes all of them to the user file, and every value applies when a
// session next opens, so the footer is the only control that reaches the kernel.
const MIN_ROUNDS = 1;
const MAX_ROUNDS = 12;
const MIN_TIMEOUT = 15;
const MAX_TIMEOUT = 600;
const MAX_CONSULTATIONS = 20;

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
      applySaved(value);
    }).catch((err: unknown) => {
      if (current) setError(reason(err));
    });
    return () => { current = false; };
  }, [port]);

  function applySaved(value: Settings) {
    setSettings(value);
    setSaved(value);
    setToolsText(value.tools.join("\n"));
  }
  if (!settings) return error ? <div role="alert">{error}</div> : null;

  const tools = toolsText.split(/\r?\n/).map((tool) => tool.trim()).filter(Boolean);
  const roundsValid = inRange(settings.researchRounds, MIN_ROUNDS, MAX_ROUNDS);
  const timeoutValid = inRange(settings.consultationTimeoutSeconds, MIN_TIMEOUT, MAX_TIMEOUT);
  const limitValid = inRange(settings.limit, 1, MAX_CONSULTATIONS);
  const valid = roundsValid && timeoutValid && limitValid;
  // A draft that failed to save stays on screen: the user typed it, and a reload
  // would lose the one thing the error is about.
  const changed = settings.consultation !== saved?.consultation || settings.limit !== saved?.limit
    || settings.researchRounds !== saved?.researchRounds
    || settings.consultationTimeoutSeconds !== saved?.consultationTimeoutSeconds
    || tools.join("\n") !== saved?.tools.join("\n");

  const save = async () => {
    setBusy(true);
    setError("");
    try {
      applySaved(await port.savePlannerSettings({ ...settings, tools }));
    } catch (err) {
      setError(reason(err));
    } finally {
      setBusy(false);
    }
  };
  const discard = () => {
    if (saved) applySaved(saved);
    setError("");
  };

  return (
    <>
      <SettingRow id="planner-research-rounds" title={t("settings.planner_research_rounds")}
        desc={settings.researchRounds >= 3
          ? t("settings.planner_research_rounds_hint", { aim: settings.researchRounds - 2 })
          : t("settings.planner_research_rounds_no_reserve")}>
        <input aria-label={t("settings.planner_research_rounds")} type="number" min={MIN_ROUNDS} max={MAX_ROUNDS}
          value={settings.researchRounds} data-action="planner.research-rounds" disabled={busy}
          aria-invalid={!roundsValid}
          onChange={(e) => setSettings({ ...settings, researchRounds: Number(e.target.value) })} />
        <Unit label={t("settings.planner_research_rounds_unit")} />
      </SettingRow>
      <SettingRow id="planner-tools" title={t("settings.planner_tools")} desc={t("settings.planner_tools_hint")}>
        <textarea aria-label={t("settings.planner_tools")} rows={3} value={toolsText} data-action="planner.tools" disabled={busy}
          onChange={(e) => setToolsText(e.target.value)} />
      </SettingRow>
      <SettingRow id="planner-consultation" title={t("settings.planner_consultation")} desc={t("settings.planner_consultation_hint")}>
        <input aria-label={t("settings.planner_consultation")} type="checkbox" data-action="planner.consultation" checked={settings.consultation} disabled={busy}
          onChange={(e) => setSettings({ ...settings, consultation: e.target.checked })} />
      </SettingRow>
      <SettingRow id="planner-consultation-limit" title={t("settings.planner_consultation_limit")}>
        <input aria-label={t("settings.planner_consultation_limit")} type="number" min={1} max={MAX_CONSULTATIONS} value={settings.limit} data-action="planner.consultation-limit"
          disabled={busy || !settings.consultation} aria-invalid={!limitValid}
          onChange={(e) => setSettings({ ...settings, limit: Number(e.target.value) })} />
        <Unit label={t("settings.planner_consultation_limit_unit")} />
      </SettingRow>
      <SettingRow id="planner-consultation-timeout" title={t("settings.planner_consultation_timeout")} desc={t("settings.planner_consultation_timeout_hint")}>
        <input aria-label={t("settings.planner_consultation_timeout")} type="number" min={MIN_TIMEOUT} max={MAX_TIMEOUT}
          value={settings.consultationTimeoutSeconds} data-action="planner.consultation-timeout"
          disabled={busy || !settings.consultation} aria-invalid={!timeoutValid}
          onChange={(e) => setSettings({ ...settings, consultationTimeoutSeconds: Number(e.target.value) })} />
        <Unit label={t("common.s")} />
      </SettingRow>
      {!valid && (
        <div className="why" role="status" data-action="planner.invalid">
          {invalidMessage(roundsValid, timeoutValid, limitValid)}
        </div>
      )}
      <div className="sfoot" data-action="planner.footer">
        <PillButton action="planner.discard" disabled={busy || !changed} onClick={discard}>
          {t("common.revert")}
        </PillButton>
        <PillButton action="planner.save" disabled={busy || !changed || !valid} onClick={() => void save()}>
          {t("common.save")}
        </PillButton>
      </div>
      {error && <div className="why" role="alert">{error}</div>}
    </>
  );
}

function Unit({ label }: { label: string }) {
  return <span className="unit" aria-hidden="true">{label}</span>;
}

function inRange(value: number, min: number, max: number) {
  return Number.isInteger(value) && value >= min && value <= max;
}

// One line naming the ranges that are actually wrong, so a user fixes the field
// they mistyped instead of re-reading every control.
function invalidMessage(rounds: boolean, timeout: boolean, limit: boolean) {
  const parts: string[] = [];
  if (!rounds) parts.push(`${MIN_ROUNDS}–${MAX_ROUNDS}`);
  if (!timeout) parts.push(`${MIN_TIMEOUT}–${MAX_TIMEOUT}`);
  if (!limit) parts.push(`1–${MAX_CONSULTATIONS}`);
  return parts.length === 0 ? "" : `${t("settings.planner_range_required", { ranges: parts.join(" · ") })}`;
}
