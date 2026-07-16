/**
 * Shared fit-to-view options for the canvas. Cap zoom at 1:1 so small graphs
 * don't render oversized, and leave a comfortable padding margin.
 */
export const FIT_OPTIONS = { maxZoom: 1, padding: 0.25 } as const;
