import { describe, it, expect } from "vitest";
import {
  entryKey,
  replaceStack,
  pushEntry,
  truncateTo,
  popEntry,
  type DrawerEntry,
} from "../drawer-stack";

const r = (index: number): DrawerEntry => ({ kind: "resource", index });
const v = (name: string): DrawerEntry => ({ kind: "volume", name });

describe("drawer-stack", () => {
  it("entryKey distinguishes kinds and identities", () => {
    expect(entryKey(r(0))).not.toBe(entryKey(r(1)));
    expect(entryKey(r(0))).not.toBe(entryKey(v("0")));
    expect(entryKey(v("data"))).toBe(entryKey(v("data")));
  });

  it("replaceStack yields a single-entry stack", () => {
    expect(replaceStack(r(2))).toEqual([r(2)]);
  });

  it("pushEntry appends a new entry", () => {
    expect(pushEntry([r(0)], v("data"))).toEqual([r(0), v("data")]);
  });

  it("pushEntry truncates to an existing entry instead of duplicating", () => {
    const stack = [r(0), v("data"), r(1)];
    expect(pushEntry(stack, v("data"))).toEqual([r(0), v("data")]);
    expect(pushEntry(stack, r(0))).toEqual([r(0)]);
  });

  it("truncateTo keeps entries up to and including depth", () => {
    const stack = [r(0), v("data"), r(1)];
    expect(truncateTo(stack, 1)).toEqual([r(0), v("data")]);
    expect(truncateTo(stack, 0)).toEqual([r(0)]);
  });

  it("popEntry removes the front entry; empty stays empty", () => {
    expect(popEntry([r(0), v("data")])).toEqual([r(0)]);
    expect(popEntry([])).toEqual([]);
  });
});
