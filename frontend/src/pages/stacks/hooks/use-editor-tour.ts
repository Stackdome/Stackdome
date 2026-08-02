import { useEffect } from "react";
import { EDITOR_TABS, type EditorTabId } from "@/pages/stacks/components/editor/editor-tabs";
import { runCanvasTour, runTimelineStep, runLiveStep, tourStage } from "@/pages/stacks/lib/onboarding/tour";

/** Drives the in-editor tour beats off editor state. The delays let the target
    DOM render before a beat tries to anchor to it. */
export function useEditorTour({
  isNewStack,
  activeTab,
  hasEndpoints,
}: {
  isNewStack: boolean;
  activeTab: EditorTabId;
  hasEndpoints: boolean;
}) {
  useEffect(() => {
    if (tourStage() !== "canvas" || !isNewStack) return;
    const t = setTimeout(runCanvasTour, 600);
    return () => clearTimeout(t);
  }, [isNewStack]);

  // Deploy was clicked: the editor switches itself to the deployments tab.
  useEffect(() => {
    if (tourStage() !== "canvas" || activeTab !== EDITOR_TABS.deployments) return;
    const t = setTimeout(runTimelineStep, 600);
    return () => clearTimeout(t);
  }, [activeTab]);

  useEffect(() => {
    if (tourStage() !== "deploying" || !hasEndpoints) return;
    const t = setTimeout(runLiveStep, 400);
    return () => clearTimeout(t);
  }, [hasEndpoints]);
}
