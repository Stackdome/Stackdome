/**
 * Monotonic guard for concurrent stack refetches. Multiple writers (autosave
 * refetch, release-transition refetch, revert, volume delete) GET the stack
 * concurrently; without ordering, a GET issued before an edit landed can
 * resolve last and clobber fresher state. Callers take a ticket via begin()
 * BEFORE issuing the request and apply the response only if shouldApply(ticket)
 * — which admits each ticket at most once and refuses any ticket older than
 * the newest applied.
 */
export interface StackFetchGate {
  begin(): number;
  shouldApply(ticket: number): boolean;
}

export function createStackFetchGate(): StackFetchGate {
  let next = 0;
  let lastApplied = 0;
  return {
    begin() {
      next += 1;
      return next;
    },
    shouldApply(ticket: number) {
      if (ticket <= lastApplied) return false;
      lastApplied = ticket;
      return true;
    },
  };
}
