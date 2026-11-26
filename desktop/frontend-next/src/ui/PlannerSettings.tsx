import { useEffect, useState } from "react";
import { t } from "../i18n";
import { reason } from "../i18n/kernel";
import type { PlannerSettings as Settings, PlannerSettingsSnapshot } from "../port/boundary";
import type { McpEntry } from "../port/mcp";
import type { ModelEntry } from "../port/model";
import type { AgentPort } from "../port/port";
import { PlannerResearchTools, plannerMcpServers } from "./PlannerResearchTools";
import { usePresetLeave } from "./PresetLeave";
import { SectionCard } from "./SectionCard";
import { PillButton, SettingRow } from "./SettingRow";
import { protocolDetail } from "./vendors";

// The research cap and the timeout are two numbers beside two switches. One
// save writes all of them to the user file, and every value applies when a
// session next opens, so the footer is the only control that reaches the kernel.
const MIN_ROUNDS = 1;
const MAX_ROUNDS = 12;
const MIN_TIMEOUT = 15;
const MAX_TIMEOUT = 600;
const MAX_CONSULTATIONS = 20;

export function PlannerSettings({ port, models = [] }: { port: AgentPort; models?: readonly ModelEntry[] }) {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [saved, setSaved] = useState<PlannerSettingsSnapshot | null>(null);
  const [custom, setCustom] = useState(false);
  const [customTools, setCustomTools] = useState<string[]>([]);
  const [servers, setServers] = useState<McpEntry[]>([]);
  const [catalogError, setCatalogError] = useState("");
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
    void port.mcp().then((catalog) => {
      if (current) setServers(catalog.servers);
    }).catch((err: unknown) => {
      if (current) setCatalogError(reason(err));
    });
    return () => { current = false; };
  }, [port]);

  function applySaved(value: PlannerSettingsSnapshot) {
    const { consultation, limit, tools, researchRounds, consultationTimeoutSeconds } = value;
    setSettings({ consultation, limit, tools, researchRounds, consultationTimeoutSeconds });
    setSaved(value);
    setCustom(value.tools.length > 0);
    setCustomTools(value.tools);
  }

  const tools = custom ? customTools : [];
  const changed = settings !== null && (settings.consultation !== saved?.consultation || settings.limit !== saved?.limit
    || settings.researchRounds !== saved?.researchRounds
    || settings.consultationTimeoutSeconds !== saved?.consultationTimeoutSeconds
    || tools.join("\n") !== saved?.tools.join("\n") || (custom && tools.length === 0));
  const { confirmation } = usePresetLeave(changed, busy);
  if (!settings) return error ? <div role="alert">{error}</div> : null;

  const roundsValid = inRange(settings.researchRounds, MIN_ROUNDS, MAX_ROUNDS);
  const timeoutValid = inRange(settings.consultationTimeoutSeconds, MIN_TIMEOUT, MAX_TIMEOUT);
  const limitValid = inRange(settings.limit, 1, MAX_CONSULTATIONS);
  const toolsValid = !custom || tools.length > 0;
  const valid = roundsValid && timeoutValid && limitValid && toolsValid;
  const builtins = saved?.researchTools ?? [];
  const mcp = plannerMcpServers(servers);
  const modelRef = saved?.modelRef;
  const model = models.find((entry) => entry.ref === modelRef);
  const modelLabel = model?.modelAlias || model?.model || modelRef;
  const protocol = model?.kind ? protocolDetail(model.kind) : "";

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
  const pickCustom = (next: boolean) => {
    if (next && !custom && customTools.length === 0) {
      setCustomTools([...builtins, ...mcp.flatMap((server) => server.tools.map((item) => item.id))]);
    }
    setCustom(next);
  };

  return (
    <>
      <SectionCard title={t("settings.planner_model_summary")}>
        <SettingRow id="planner-model" title={t("settings.planner_effective_model")} desc={t("settings.planner_current_session")}>
          <span className="planner-model-value">{[modelLabel, protocol].filter(Boolean).join(" · ") || t("settings.planner_no_model")}</span>
        </SettingRow>
      </SectionCard>
      <SectionCard title={t("settings.planner_research_section")}>
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
      <PlannerResearchTools builtins={builtins} servers={mcp} selected={customTools} custom={custom}
        busy={busy} catalogError={catalogError} onMode={pickCustom} onChange={setCustomTools} />
      </SectionCard>
      <SectionCard title={t("settings.planner_consultations_section")}>
      <SettingRow id="planner-consultation" title={t("settings.planner_consultation")} desc={t("settings.planner_consultation_hint")}>
        <input className="toggle" aria-label={t("settings.planner_consultation")} type="checkbox" data-action="planner.consultation" checked={settings.consultation} disabled={busy}
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
      </SectionCard>
      {(!roundsValid || !timeoutValid || !limitValid) && (
        <div className="why" role="status" data-action="planner.invalid">
          {invalidMessage(roundsValid, timeoutValid, limitValid)}
        </div>
      )}
      {!toolsValid && <div className="why" role="status">{t("settings.planner_custom_empty")}</div>}
      {error && <div className="why" role="alert">{error}</div>}
      <div className="planner-footer" data-action="planner.footer">
        <span>{t("settings.applies_to_new_sessions")}</span>
        <PillButton action="planner.discard" disabled={busy || !changed} onClick={discard}>
          {t("common.revert")}
        </PillButton>
        <button type="button" className="btn sm" data-primary="" data-action="planner.save"
          disabled={busy || !changed || !valid} onClick={() => void save()}>
          {t("common.save")}
        </button>
      </div>
      {confirmation}
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
