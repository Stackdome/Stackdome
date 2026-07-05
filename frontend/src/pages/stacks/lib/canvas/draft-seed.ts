import type {
  FormStackData,
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

/** A create-flow seed handed to the draft canvas via navigation state. */
export interface DraftSeed {
  name: string;
  labels: FormStackData["labels"];
  resources: FormStackResourceData[];
  volumes: FormVolumeExtendedData[];
  linkedAddonIds: string[];
}

export function emptyDraftSeed(): DraftSeed {
  return { name: "", labels: [], resources: [], volumes: [], linkedAddonIds: [] };
}

/** Assemble the validated create payload shape from the canvas draft state. */
export function buildDraftFormData(
  name: string,
  labels: FormStackData["labels"],
  resources: FormStackResourceData[],
  volumes: FormVolumeExtendedData[],
): FormStackData {
  return {
    name,
    labels,
    spec: { stack_resources: resources, volumes },
  } as FormStackData;
}
