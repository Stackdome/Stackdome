import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogClose } from "@/components/ui/dialog";
import ClusterCreateForm from "../create/index";
import { PlusCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import * as api from "@/api/clusters";
import { getCurrentOrganizationId } from "@/helpers/common";

interface ClusterCreationModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function ClusterCreationModal({ open, onOpenChange, onSuccess }: ClusterCreationModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(values: api.Cluster) {
    setLoading(true);
    setError(null);
    try {
      await api.createCluster(getCurrentOrganizationId(), values);
      setLoading(false);
      onSuccess();
      onOpenChange(false);
    } catch (e: unknown) {
      console.error("Failed to create cluster:", e);
      if (e instanceof Error) {
        setError(e.message);
      } else if (typeof e === 'object' && e !== null && 'message' in e) {
        setError(String(e.message));
      } else {
        setError("Failed to create cluster. Please check your connection and try again.");
      }
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl w-full max-h-[90vh] h-auto overflow-y-auto">
        <DialogHeader className="pb-4 border-b">
          <DialogTitle className="text-xl font-semibold flex items-center gap-2">
            <PlusCircle className="h-5 w-5 text-primary" />
            Create New Cluster
          </DialogTitle>
        </DialogHeader>
        <div className="py-6 px-1">
          <ClusterCreateForm onSubmit={handleSubmit} loading={loading} error={error} />
        </div>
        <DialogClose asChild>
          <Button variant="ghost" size="icon" className="absolute top-4 right-4">
            <span className="sr-only">Close</span>
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </Button>
        </DialogClose>
      </DialogContent>
    </Dialog>
  );
}
