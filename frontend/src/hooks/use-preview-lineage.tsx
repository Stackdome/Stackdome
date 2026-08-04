import { useContext } from "react";
import { PreviewLineageContext } from "@/contexts/preview-lineage-context";

export function usePreviewLineage() {
  return useContext(PreviewLineageContext);
}
