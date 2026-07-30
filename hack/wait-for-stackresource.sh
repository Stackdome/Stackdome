#!/usr/bin/env bash
# Wait until the cluster agent has reconciled a StackResource at or past a given
# generation and reports Converged=True.
#
# Why not `kubectl wait --for=condition=Converged=True`: that ignores
# observedGeneration and matches the condition left over from the previous spec,
# returning success before the agent has looked at the new image at all.
#
# Convergence is read from the Converged condition's OWN observedGeneration,
# which is the same test the agent itself applies (isResourceConverged:
# cond.Status == True && cond.ObservedGeneration == resource.Generation). Both
# halves come out of one API read, so there is no window where a stale True
# pairs with a fresh generation.
set -uo pipefail

NAME=${1:?usage: wait-for-stackresource.sh <name> <namespace> <min-generation> [timeout-seconds]}
NAMESPACE=${2:?namespace required}
MIN_GEN=${3:?min-generation required}
TIMEOUT=${4:-600}
INTERVAL=5

case $MIN_GEN in
  '' | *[!0-9]*) echo "error: min-generation must be a non-negative integer, got: '$MIN_GEN'" >&2; exit 1 ;;
esac
case $TIMEOUT in
  '' | *[!0-9]*) echo "error: timeout-seconds must be a non-negative integer, got: '$TIMEOUT'" >&2; exit 1 ;;
esac

kget() { kubectl get stackresource "$NAME" -n "$NAMESPACE" "$@" 2>/dev/null; }

# Reads one condition's status and its own observedGeneration in a single API
# call, as "<status>|<generation>". A condition that does not exist yields an
# empty string rather than a shifted field, so parsing stays safe.
cond_pair() {
  kget -o jsonpath="{range .status.conditions[?(@.type==\"$1\")]}{.status}|{.observedGeneration}{end}"
}

dump() {
  echo "--- conditions ---"
  kget -o jsonpath='{range .status.conditions[*]}{.type}={.status} gen={.observedGeneration} reason={.reason} message={.message}{"\n"}{end}'
  echo "--- lastFailureDetail ---"
  kget -o jsonpath='{range .status.lastFailureDetail[*]}{.containerName}: {.lastTerminationReason} exit={.lastTerminationExitCode} restarts={.restartCount} message={.lastTerminationMessage}{"\n"}{end}'
  echo "--- pods ---"
  kubectl get pods -n "$NAMESPACE" -l "resource=$NAME" -o wide
  echo "--- logs (current) ---"
  kubectl logs -n "$NAMESPACE" -l "resource=$NAME" --all-containers --tail=100
  echo "--- logs (previous) ---"
  kubectl logs -n "$NAMESPACE" -l "resource=$NAME" --all-containers --previous --tail=100
}

# Preflight: stderr is deliberately NOT suppressed here. Without it a wrong
# name, wrong namespace, rotated token, unreachable API server or missing CRD
# looks identical to "not reconciled yet" and burns the whole timeout in silence.
if ! kubectl get stackresource "$NAME" -n "$NAMESPACE" -o name; then
  echo "preflight failed: cannot read stackresource $NAME in namespace $NAMESPACE" >&2
  exit 1
fi

deadline=$(( $(date +%s) + TIMEOUT ))

while :; do
  snapshot=$(kget -o jsonpath='{.status.observedGeneration}|{.status.phase}')
  observed=${snapshot%%|*}
  phase=${snapshot##*|}
  observed=${observed:-0}
  phase=${phase:-Unknown}

  conv=$(cond_pair Converged)
  conv_status=${conv%%|*}
  conv_gen=${conv##*|}
  conv_status=${conv_status:-Unknown}
  conv_gen=${conv_gen:-0}

  stalled=$(cond_pair Stalled)
  stalled_status=${stalled%%|*}
  stalled_gen=${stalled##*|}
  stalled_status=${stalled_status:-Unknown}
  stalled_gen=${stalled_gen:-0}

  if [ "$conv_gen" -ge "$MIN_GEN" ] && [ "$conv_status" = "True" ]; then
    echo "converged: Converged=True at observedGeneration=$conv_gen (>= $MIN_GEN) phase=$phase"
    exit 0
  fi

  # Only trust failure signals once the agent has stamped them with our
  # generation, otherwise a stale failure from the previous spec aborts a
  # healthy deploy. Stalled=True is the agent's terminal-failure marker and,
  # unlike phase, carries its own generation. Degraded/Pending are NOT failures
  # -- Degraded is the normal mid-rollout state.
  if [ "$stalled_gen" -ge "$MIN_GEN" ] && [ "$stalled_status" = "True" ]; then
    echo "FAILED: Stalled=True at observedGeneration=$stalled_gen (>= $MIN_GEN) phase=$phase"
    dump
    exit 1
  fi

  if [ "$observed" -ge "$MIN_GEN" ] && [ "$phase" = "Failed" ]; then
    echo "FAILED at observedGeneration=$observed phase=$phase"
    dump
    exit 1
  fi

  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "TIMEOUT after ${TIMEOUT}s: Converged=$conv_status@gen$conv_gen (want True@gen>=$MIN_GEN) observedGeneration=$observed phase=$phase"
    dump
    exit 1
  fi

  echo "waiting: Converged=$conv_status@gen$conv_gen/$MIN_GEN observedGeneration=$observed phase=$phase stalled=$stalled_status@gen$stalled_gen"
  sleep "$INTERVAL"
done
