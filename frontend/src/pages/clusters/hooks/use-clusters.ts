import { useCallback, useEffect, useState } from "react";
import { getCurrentOrganizationId } from "@/helpers/common";
import { z } from "zod";
import type { Cluster } from "../types";
import { getErrorMessage, isNotFoundError } from "@/api/client";
import * as clusterApi from "@/api/clusters";

export const ClusterSchema = z.object({
  name: z.string().min(1, "Cluster name is required"),
  cluster_url: z.string().url("API server URL must be a valid URL"),
  cluster_ca_data: z.string().min(1, "CA cert is required"),
  cluster_sa_token: z.string().min(1, "Service account token is required"),
  organisation_id: z.string().optional(),
  id: z.string().optional(),
  default: z.boolean().optional(),
  cluster_image_registry: z.object({
    name: z.string().min(1, "Image registry name is required"),
    spec: z.object({
      backend_storage_size: z.string().min(1, "Registry size is required"),
    }),
  }).optional(),
});

export type ClusterData = z.infer<typeof ClusterSchema>;

function mapClusterFormInputToApiPayload(input: ClusterData): Omit<Cluster, "id"> {
  const { cluster_image_registry, ...rest } = input;
  return {
    ...rest,
    ...(cluster_image_registry ? { cluster_image_registry: { name: "default-registry" } } : {}),
  };
}

export function useClusters() {
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const orgId = getCurrentOrganizationId();

  const fetchClusters = useCallback(async () => {
    if (!orgId) return;
    setLoading(true);
    setError(null);
    try {
      const data = await clusterApi.getClusters(orgId);
      setClusters(data.items || []);
    } catch (e: unknown) {
      if (isNotFoundError(e)) {
        // Treat "not found" as empty state, not an error
        setClusters([]);
      } else {
        setError(getErrorMessage(e));
      }
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  useEffect(() => {
    fetchClusters();
  }, [fetchClusters]);

  return { clusters, loading, error, refetch: fetchClusters };
}

export function useCreateCluster() {
  const orgId = getCurrentOrganizationId();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<Cluster | null>(null);

  const createCluster = useCallback(async (input: ClusterData) => {
    if (!orgId) throw new Error("No organization selected");
    setLoading(true);
    setError(null);
    setData(null);
    try {
      const parsed = ClusterSchema.parse({ ...input, organisation_id: orgId });
      const apiPayload = mapClusterFormInputToApiPayload(parsed);
      const cluster = await clusterApi.createCluster(orgId, apiPayload);
      setData(cluster);
      return cluster;
    } catch (e) {
      setError(getErrorMessage(e));
      throw e;
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  return { createCluster, loading, error, data };
}

export function useDeleteCluster() {
  const orgId = getCurrentOrganizationId();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const deleteCluster = useCallback(async (clusterId: string) => {
    if (!orgId) throw new Error("No organization selected");
    setLoading(true);
    setError(null);
    setSuccess(false);
    try {
      await clusterApi.deleteCluster(orgId, clusterId);
      setSuccess(true);
      return true;
    } catch (e) {
      setError(getErrorMessage(e));
      throw e;
    } finally {
      setLoading(false);
    }
  }, [orgId]);

  return { deleteCluster, loading, error, success };
}
