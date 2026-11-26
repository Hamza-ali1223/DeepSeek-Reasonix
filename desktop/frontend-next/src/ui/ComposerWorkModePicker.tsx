import { t } from "../i18n";
import { Picker } from "./Menu";
import { StudioIcon } from "./StudioIcon";

export function ComposerWorkModePicker({ mode, pending, onPick }: {
  mode: "agent" | "plan" | "agent_planner";
  pending?: boolean;
  onPick: (value: string) => void;
}) {
  return (
    <div className="studio-mode-control">
      <Picker
        className="mode studio-mode-picker"
        triggerAction="plan.mode"
        data-action="plan.mode"
        ariaPressed={mode === "plan"}
        place="bottom"
        title={t("composer.work_mode")}
        menuClassName="composer-menu mode-menu"
        current={mode}
        pending={pending}
        items={[
          { value: "__mode", label: t("composer.work_mode"), right: t("common.current_task"), header: true },
          { value: "agent", label: "Agent", desc: t("composer.run_tasks_with_enabled_tools") },
          { value: "plan", label: "Plan", desc: t("composer.plan_steps_before_changing_files") },
          { value: "agent_planner", label: t("composer.agent_plus_planner"), desc: t("composer.plan_with_dedicated_model_before_execution") },
        ]}
        onPick={onPick}
        label={<><StudioIcon name={mode === "plan" ? "tasks" : "agent"} /><span className="studio-sr-label">{t("common.plan")}</span><span>{mode === "agent_planner" ? t("composer.agent_plus_planner") : mode === "plan" ? "Plan" : "Agent"}</span><StudioIcon name="down" /></>}
      />
    </div>
  );
}
