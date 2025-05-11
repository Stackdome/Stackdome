import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogClose } from "@/components/ui/dialog";
import ClusterCreateForm from "./index";
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
      await new Promise(res => setTimeout(res, 1000));
      setLoading(false);
      onSuccess();
      onOpenChange(false);
    } catch (e: unknown) {
      if (e instanceof Error) {
        setError(e.message);
      } else {
        setError("Failed to create cluster");
      }
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg w-full h-screen flex flex-col justify-center bg-background">
        <DialogHeader>
          <DialogTitle className="text-2xl font-bold flex items-center gap-2">
            <PlusCircle className="h-6 w-6" />
            Create New Cluster
          </DialogTitle>
        </DialogHeader>
        <div className="py-4">
          <ClusterCreateForm onSubmit={handleSubmit} loading={loading} error={error} />
        </div>
        <DialogClose asChild>
          <Button variant="ghost" className="absolute top-4 right-4">Close</Button>
        </DialogClose>
      </DialogContent>
    </Dialog>
  );
}
