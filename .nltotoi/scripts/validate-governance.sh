#!/bin/bash
# validate-governance.sh — Check that all required governance files exist
# per ORG-DEV-OTOI-1.0.3.
#
# Usage: bash .nltotoi/scripts/validate-governance.sh
# Exit code: 0 if compliant, 1 if any required file is missing.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

REQUIRED_FILES=(
  "CLAUDE.md"
  "docs/active-threads.md"
  "docs/agent-log/README.md"
  "templates/agent-registration.json"
  "templates/handoff-record.json"
  "templates/escalation.md"
  "templates/intent-log.md"
  "SOPs/new-agent-onboarding.md"
  "SOPs/repo-governance-setup.md"
  "SOPs/incident-response.md"
  "ISSUE_TEMPLATE/agent-escalation.md"
  "ISSUE_TEMPLATE/governance-proposal.md"
  ".nltotoi/index/governance-files.md"
  ".nltotoi/scripts/validate-governance.sh"
)

REQUIRED_DIRS=(
  "docs/agent-log/registrations"
  "docs/agent-log/handoffs"
)

FAILED=0

echo "=== NLT Governance Validation (ORG-DEV-OTOI-1.0.3) ==="
echo "Repo: $REPO_ROOT"
echo ""

for f in "${REQUIRED_FILES[@]}"; do
  if [[ -f "$REPO_ROOT/$f" ]]; then
    echo "  ✅ $f"
  else
    echo "  ❌ MISSING: $f"
    FAILED=1
  fi
done

for d in "${REQUIRED_DIRS[@]}"; do
  if [[ -d "$REPO_ROOT/$d" ]]; then
    echo "  ✅ $d/"
  else
    echo "  ❌ MISSING DIR: $d/"
    FAILED=1
  fi
done

echo ""
if [[ $FAILED -eq 0 ]]; then
  echo "✅ All governance files present. Repo is compliant."
else
  echo "❌ Some governance files are missing. See above."
fi

exit $FAILED
