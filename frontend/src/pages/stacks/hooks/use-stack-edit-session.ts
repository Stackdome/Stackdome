import { useCallback, useMemo, useState } from "react";
import {
  cloneJson,
  diffStack,
  revertResource,
  revertVolume,
  type StackDiff,
  type ResourceArr,
  type VolumeArr,
} from "@/pages/stacks/lib/stack-diff";

export type EditSessionTab = "configuration" | "deployment" | "environment";

export interface EditSessionDraft {
  resources: ResourceArr;
  volumes: VolumeArr;
}

export interface EditSessionState {
  isActive: boolean;
  draft: EditSessionDraft;
  baseline: EditSessionDraft;
  openResourceIdx: number | null;
  openVolumeIdx: number | null;
  openTab: EditSessionTab | null;
}

export interface EditSessionStartOpts {
  openResourceIdx?: number | null;
  openVolumeIdx?: number | null;
  openTab?: EditSessionTab | null;
}

const EMPTY_DRAFT: EditSessionDraft = { resources: [], volumes: [] };

const INITIAL_STATE: EditSessionState = {
  isActive: false,
  draft: EMPTY_DRAFT,
  baseline: EMPTY_DRAFT,
  openResourceIdx: null,
  openVolumeIdx: null,
  openTab: null,
};

export interface UseStackEditSession {
  isActive: boolean;
  draft: EditSessionDraft;
  baseline: EditSessionDraft;
  openResourceIdx: number | null;
  openVolumeIdx: number | null;
  openTab: EditSessionTab | null;
  dirty: StackDiff;
  linkedAddonIds: Set<string>;
  pendingDetach: Set<string>;
  start: (baseline: EditSessionDraft, opts?: EditSessionStartOpts) => void;
  discard: () => void;
  discardResource: (idx: number) => void;
  discardVolume: (idx: number) => void;
  updateResources: (
    updater:
      | ResourceArr
      | ((prev: ResourceArr) => ResourceArr),
  ) => void;
  updateVolumes: (
    updater:
      | VolumeArr
      | ((prev: VolumeArr) => VolumeArr),
  ) => void;
  setLinkedAddonIds: (next: Set<string> | ((prev: Set<string>) => Set<string>)) => void;
  setPendingDetach: (next: Set<string> | ((prev: Set<string>) => Set<string>)) => void;
}

export function useStackEditSession(): UseStackEditSession {
  const [state, setState] = useState<EditSessionState>(INITIAL_STATE);
  const [linkedAddonIds, setLinkedAddonIdsState] = useState<Set<string>>(new Set());
  const [pendingDetach, setPendingDetachState] = useState<Set<string>>(new Set());

  const start = useCallback(
    (baseline: EditSessionDraft, opts?: EditSessionStartOpts) => {
      const cloned: EditSessionDraft = {
        resources: cloneJson(baseline.resources),
        volumes: cloneJson(baseline.volumes),
      };
      const baselineSnap: EditSessionDraft = {
        resources: cloneJson(baseline.resources),
        volumes: cloneJson(baseline.volumes),
      };
      setState({
        isActive: true,
        draft: cloned,
        baseline: baselineSnap,
        openResourceIdx: opts?.openResourceIdx ?? null,
        openVolumeIdx: opts?.openVolumeIdx ?? null,
        openTab: opts?.openTab ?? null,
      });
      setLinkedAddonIdsState(new Set());
      setPendingDetachState(new Set());
    },
    [],
  );

  const discard = useCallback(() => {
    setState(INITIAL_STATE);
    setLinkedAddonIdsState(new Set());
    setPendingDetachState(new Set());
  }, []);

  const discardResource = useCallback((idx: number) => {
    setState((prev) => {
      if (!prev.isActive) return prev;
      const nextDraft = revertResource(prev.draft, prev.baseline, idx);
      return { ...prev, draft: nextDraft };
    });
  }, []);

  const discardVolume = useCallback((idx: number) => {
    setState((prev) => {
      if (!prev.isActive) return prev;
      const nextDraft = revertVolume(prev.draft, prev.baseline, idx);
      return { ...prev, draft: nextDraft };
    });
  }, []);

  const updateResources = useCallback(
    (updater: ResourceArr | ((prev: ResourceArr) => ResourceArr)) => {
      setState((prev) => {
        if (!prev.isActive) return prev;
        const nextResources =
          typeof updater === "function" ? updater(prev.draft.resources) : updater;
        return { ...prev, draft: { ...prev.draft, resources: nextResources } };
      });
    },
    [],
  );

  const updateVolumes = useCallback(
    (updater: VolumeArr | ((prev: VolumeArr) => VolumeArr)) => {
      setState((prev) => {
        if (!prev.isActive) return prev;
        const nextVolumes =
          typeof updater === "function" ? updater(prev.draft.volumes) : updater;
        return { ...prev, draft: { ...prev.draft, volumes: nextVolumes } };
      });
    },
    [],
  );

  const setLinkedAddonIds = useCallback(
    (next: Set<string> | ((prev: Set<string>) => Set<string>)) => {
      setLinkedAddonIdsState((prev) =>
        typeof next === "function" ? (next as (p: Set<string>) => Set<string>)(prev) : next,
      );
    },
    [],
  );

  const setPendingDetach = useCallback(
    (next: Set<string> | ((prev: Set<string>) => Set<string>)) => {
      setPendingDetachState((prev) =>
        typeof next === "function" ? (next as (p: Set<string>) => Set<string>)(prev) : next,
      );
    },
    [],
  );

  const dirty = useMemo<StackDiff>(
    () => diffStack(state.draft, state.baseline),
    [state.draft, state.baseline],
  );

  return {
    isActive: state.isActive,
    draft: state.draft,
    baseline: state.baseline,
    openResourceIdx: state.openResourceIdx,
    openVolumeIdx: state.openVolumeIdx,
    openTab: state.openTab,
    dirty,
    linkedAddonIds,
    pendingDetach,
    start,
    discard,
    discardResource,
    discardVolume,
    updateResources,
    updateVolumes,
    setLinkedAddonIds,
    setPendingDetach,
  };
}
