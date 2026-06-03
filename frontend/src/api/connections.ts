// Type re-exports for the connections schema — consumed by connection-mapping.ts.
// The CRUD client (listStackConnections, createStackConnection, etc.) was removed
// after the save path moved to a full connection set in the stack PUT spec.
import type { components } from "./types/openapi";

export type StackConnection = components["schemas"]["StackConnection"];
export type StackConnectionList = components["schemas"]["StackConnectionList"];
export type ConnectionMapping = components["schemas"]["ConnectionMapping"];
export type TopologyNodeRef = components["schemas"]["TopologyNodeRef"];
export type StackConnectionConfig = components["schemas"]["StackConnectionConfig"];
