import { useMemo, useState } from "react";
import { t } from "../i18n";
import type { McpEntry } from "../port/mcp";
import { SegChoice } from "./SettingRow";

type Server = { name: string; tools: { id: string; name: string; description?: string }[] };

export function plannerMcpServers(entries: readonly McpEntry[]): Server[] {
  return entries.filter((entry) => entry.enabled && entry.state === "ready" && !entry.remembered && !entry.stale && !entry.error)
    .map((entry) => ({
      name: entry.name,
      tools: (entry.toolList ?? []).filter((tool) => tool.readOnly && !tool.destructive && !tool.error)
        .map((tool) => ({ id: `mcp-tool:${entry.name}/${tool.name}`, name: tool.name, description: tool.description })),
    })).filter((entry) => entry.tools.length > 0);
}

export function PlannerResearchTools({ builtins, servers, selected, custom, busy, catalogError, onMode, onChange }: {
  builtins: readonly string[]; servers: readonly Server[]; selected: readonly string[]; custom: boolean; busy: boolean;
  catalogError: string; onMode: (custom: boolean) => void; onChange: (next: string[]) => void;
}) {
  const [search, setSearch] = useState("");
  const available = useMemo(() => new Set([...builtins, ...servers.flatMap((server) => server.tools.map((tool) => tool.id))]), [builtins, servers]);
  const unknown = selected.filter((id) => !available.has(id));
  const query = search.trim().toLowerCase();
  const matches = (name: string, description?: string) => !query || `${name} ${description ?? ""}`.toLowerCase().includes(query);
  const toggle = (id: string, checked: boolean) => onChange(checked ? [...new Set([...selected, id])] : selected.filter((item) => item !== id));

  return <div className="planner-tools">
    <div className="planner-tools-top">
      <SegChoice label={t("settings.planner_tools_mode")} value={custom ? "custom" : "default"}
        options={[["default", t("settings.planner_tools_default")], ["custom", t("settings.planner_tools_custom")]]}
        disabled={busy} action="planner.tools-mode" onPick={(value) => onMode(value === "custom")} />
    </div>
    {custom && <>
      <label className="planner-tools-search"><span className="sr-only">{t("settings.planner_tools_search")}</span>
        <input type="search" data-action="planner.tools-search" value={search} placeholder={t("settings.planner_tools_search")} disabled={busy} onChange={(event) => setSearch(event.target.value)} />
      </label>
      {catalogError && <p className="planner-tools-note" role="status">{t("settings.planner_tools_catalog_unavailable", { detail: catalogError })}</p>}
      <div className="planner-tools-list">
        <h5>{t("settings.planner_tools_builtin")}</h5>
        {builtins.filter((name) => matches(name)).map((name) => <Tool key={name} id={name} name={name} selected={selected.includes(name)} disabled={busy} onChange={toggle} />)}
        {servers.map((server) => {
          const tools = server.tools.filter((tool) => matches(tool.name, tool.description));
          return tools.length ? <section key={server.name} className="planner-tools-server"><h5>{server.name}</h5>
            {tools.map((tool) => <Tool key={tool.id} {...tool} selected={selected.includes(tool.id)} disabled={busy} onChange={toggle} />)}
          </section> : null;
        })}
        {unknown.length > 0 && <section className="planner-tools-server"><h5>{t("settings.planner_tools_unavailable")}</h5>
          {unknown.filter((id) => matches(id)).map((id) => <Tool key={id} id={id} name={id} selected disabled={busy} onChange={toggle} />)}
        </section>}
        {builtins.length === 0 && servers.length === 0 && unknown.length === 0 && <p className="planner-tools-note">{t("settings.planner_tools_none")}</p>}
      </div>
    </>}
  </div>;
}

function Tool({ id, name, description, selected, disabled, onChange }: {
  id: string; name: string; description?: string; selected: boolean; disabled: boolean;
  onChange: (id: string, checked: boolean) => void;
}) {
  return <label className="planner-tool"><input type="checkbox" data-action="planner.tools-toggle" checked={selected} disabled={disabled} onChange={(event) => onChange(id, event.target.checked)} />
    <span><strong>{name}</strong>{description && <small>{description}</small>}</span>
  </label>;
}
