import { useEffect } from "react";
import { EDITOR_TABS, type EditorTabId } from "@/pages/stacks/components/editor/editor-tabs";
import { runCanvasTour, runTimelineStep, runLiveStep, tourStage } from "./tour";

/** Drives the in-editor tour beats off editor state. Inert unless the stacks
    list started a tour this session. Delays let the target DOM render first. */
export function useEditorTour({
  isNewStack,
  activeTab,
  hasEndpoints,
}: {
  isNewStack: boolean;
  activeTab: EditorTabId;
  hasEndpoints: boolean;
}) {
  // Beats 2–5: the seeded draft canvas.
  useEffect(() => {
    if (tourStage() !== "canvas" || !isNewStack) return;
    const t = setTimeout(runCanvasTour, 600);
    return () => clearTimeout(t);
  }, [isNewStack]);

  // Beat 6: Deploy was clicked — the editor switches itself to the
  // deployments tab, which is our signal.
  useEffect(() => {
    if (tourStage() !== "canvas" || activeTab !== EDITOR_TABS.deployments) return;
    const t = setTimeout(runTimelineStep, 600);
    return () => clearTimeout(t);
  }, [activeTab]);

  // Beat 8: the release converged and a public URL exists.
  useEffect(() => {
    if (tourStage() !== "deploying" || !hasEndpoints) return;
    const t = setTimeout(runLiveStep, 400);
    return () => clearTimeout(t);
  }, [hasEndpoints]);
}
